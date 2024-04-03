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
        } else if (key === "garage:requested-state") {
            this.onGarageRequestedStateUpdate(message);
        }
    }

    onConnectionStatusUpdate(message) {
        const connectionStatus = message.getValue().unpack(
            proto.qmq.QMQConnectionState.deserializeBinary,
            "qmq.QMQConnectionState");
        
        this._vm.websocketConnected = connectionStatus;

        if  (!this._vm.serverInteractor) {
            return;
        }

        this._vm.serverInteractor.get('garage:state');
        this._vm.serverInteractor.get('garage:requested-state');
    }

    onGarageStateUpdate(message) {
        const garageState = message.getValue().unpack(
            proto.qmq.QMQGarageDoorState.deserializeBinary,
            "qmq.QMQGarageDoorState");

        this._vm.state = garageState;
    }

    onGarageRequestedStateUpdate(message) {
        const garageRequestedState = message.getValue().unpack(
            proto.qmq.QMQGarageDoorState.deserializeBinary,
            "qmq.QMQGarageDoorState");

        this._vm.requestedState = garageRequestedState;
    }
}

function NewRemoteApplication() {
    return Vue.createApp({
        data() {
            const listener = new RemoteServiceListener(this);

            return {
                websocketConnected: new proto.qmq.QMQConnectionState(),
                requestedState: new proto.qmq.QMQGarageDoorState(),
                state: new proto.qmq.QMQGarageDoorState(),
                serverInteractor:
                    new ServerInteractor(`ws://${window.location.hostname}:20000/ws`, new NotificationManager()
                        .addListener('connected', listener)
                        .addListener('garage:state', listener)
                        .addListener('garage:requested-state', listener))
            }
        },
        mounted() {
            this.serverInteractor.connect()
        },
        methods: {
            onButtonClick: function () {
                let requestedStateValue = proto.qmq.QMQGarageDoorStateEnum.GARAGE_DOOR_STATE_UNSPECIFIED;

                if (this.state.getValue() === proto.qmq.QMQGarageDoorStateEnum.GARAGE_DOOR_STATE_CLOSED) {
                    requestedStateValue = proto.qmq.QMQGarageDoorStateEnum.GARAGE_DOOR_STATE_OPEN;
                } else if (this.state.getValue() === proto.qmq.QMQGarageDoorStateEnum.GARAGE_DOOR_STATE_OPEN) {
                    requestedStateValue = proto.qmq.QMQGarageDoorStateEnum.GARAGE_DOOR_STATE_CLOSED;
                }

                this.requestedState.setValue(requestedStateValue);

                const value = new proto.google.protobuf.Any();
                value.pack(this.requestedState.serializeBinary(), 'qmq.QMQGarageDoorState');
                this.serverInteractor.set('garage:requested-state', value);
            }
        },
        computed: {
            stateAsText: function () {
                const opened = proto.qmq.QMQGarageDoorStateEnum.GARAGE_DOOR_STATE_OPEN;
                const closed = proto.qmq.QMQGarageDoorStateEnum.GARAGE_DOOR_STATE_CLOSED;
                const currentState = this.state.getValue();
                const requestedState = this.requestedState.getValue();

                if (currentState === opened && currentState === requestedState) {
                    return "Opened"
                } else if (currentState === closed && currentState === requestedState) {
                    return "Closed"
                } else if (currentState === opened && requestedState === closed) {
                    return "Closing"
                } else if (currentState === closed && requestedState === opened) {
                    return "Opening"
                } else {
                    return "Unknown"
                }
            },
            fullyConnected: function () {
                return this.websocketConnected.getValue() === proto.qmq.QMQConnectionStateEnum.CONNECTION_STATE_CONNECTED;
            }
        }
    })
}