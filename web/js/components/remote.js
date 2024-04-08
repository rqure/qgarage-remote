class RemoteServiceListener extends NotificationListener {
    constructor(vm) {
        super()
        this._vm = vm;
    }

    onNotification(key, message, context) {
        if (key === "connected") {
            this.onConnectionStatusUpdate(message);
        } else if (key === "garage:state") {
            this.onGarageStateUpdate(message);
        }
    }

    onConnectionStatusUpdate(message) {
        const connectionStatus = message.getValue().unpack(
            proto.qmq.ConnectionState.deserializeBinary,
            "qmq.ConnectionState");
        
        this._vm.websocketConnected = connectionStatus;

        if  (!this._vm.serverInteractor) {
            return;
        }

        this._vm.serverInteractor.get('garage:state');
    }

    onGarageStateUpdate(message) {
        const garageState = message.getValue().unpack(
            proto.qmq.GarageDoorState.deserializeBinary,
            "qmq.GarageDoorState");

        this._vm.state = garageState;
    }
}

function NewRemoteApplication() {
    return Vue.createApp({
        data() {
            const listener = new RemoteServiceListener(this);

            return {
                websocketConnected: new proto.qmq.ConnectionState(),
                state: new proto.qmq.GarageDoorState(),
                trigger: new proto.qmq.Int(),
                serverInteractor:
                    new ServerInteractor(`ws://${window.location.hostname}:20000/ws`, new NotificationManager()
                        .addListener('connected', listener)
                        .addListener('garage:state', listener))
            }
        },
        mounted() {
            this.serverInteractor.connect()
        },
        methods: {
            onButtonClick: function () {
                const value = new proto.google.protobuf.Any();
                value.pack(this.trigger.serializeBinary(), 'qmq.Int');
                this.serverInteractor.set('garage:trigger', value);
            }
        },
        computed: {
            stateAsText: function () {
                const opened = proto.qmq.GarageDoorState.GarageDoorStateEnum.OPEN;
                const closed = proto.qmq.GarageDoorState.GarageDoorStateEnum.CLOSED;
                const currentState = this.state.getValue();

                if (currentState === opened) {
                    return "Opened"
                } else if (currentState === closed) {
                    return "Closed"
                } else {
                    return "Unknown"
                }
            },
            fullyConnected: function () {
                return this.websocketConnected.getValue() === proto.qmq.ConnectionStateEnum.CONNECTION_STATE_CONNECTED;
            }
        }
    })
}