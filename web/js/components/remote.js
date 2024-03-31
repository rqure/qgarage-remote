const GARAGE_DOOR_STATE_UNSPECIFIED = 0;
const GARAGE_DOOR_STATE_OPENED = 1;
const GARAGE_DOOR_STATE_CLOSED = 2;

class RemoteServiceListener extends NotificationListener {
    constructor(vm) {
        super()
        this._vm = vm;
    }

    onNotification(key, data, context) {
        if (key === "connected") {
            this._vm.websocketConnected = data.value;
        } else if (key === "garage:state") {
            this._vm.state = data.value;
        } else if (key === "garage:requested-state") {
            this._vm.requestedState = data.value;
        }
    }
}

const remoteApp = Vue.createApp({
    data() {
        const listener = new RemoteServiceListener(this);

        return {
            websocketConnected: false,
            requestedState: GARAGE_DOOR_STATE_UNSPECIFIED,
            state: GARAGE_DOOR_STATE_UNSPECIFIED,
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
        onButtonClick: function() {
            if (this.state == GARAGE_DOOR_STATE_CLOSED) {
                this.serverInteractor.set('garage:requested-state', GARAGE_DOOR_STATE_OPENED)
            } else if (this.state == GARAGE_DOOR_STATE_OPENED) {
                this.serverInteractor.set('garage:requested-state', GARAGE_DOOR_STATE_CLOSED)
            }
        }
    },
    computed: {
        stateAsText: function() {
            if (this.state === GARAGE_DOOR_STATE_OPENED && this.state === this.requestedState) {
                return "Opened"
            } else if (this.state === GARAGE_DOOR_STATE_CLOSED && this.state === this.requestedState) {
                return "Closed"
            } else if (this.state === GARAGE_DOOR_STATE_OPENED && this.requestedState === GARAGE_DOOR_STATE_CLOSED) {
                return "Closing"
            } else if (this.state === GARAGE_DOOR_STATE_CLOSED && this.requestedState === GARAGE_DOOR_STATE_OPENED) {
                return "Opening"
            } else {
                return "Unknown"
            }
        },
        fullyConnected: function() {
            return this.websocketConnected;
        }
    }
})