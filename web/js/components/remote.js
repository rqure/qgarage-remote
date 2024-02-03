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
        } else if (key === "garage:shelly:connected") {
            this._vm.shellyConnected = data.value;
        }
    }
}

const remoteApp = Vue.createApp({
    data() {
        const listener = new RemoteServiceListener(this);

        return {
            websocketConnected: false,
            shellyConnected: false,
            state: "closed",
            serverInteractor:
                new ServerInteractor(`ws://${window.location.hostname}:20000/ws`, new NotificationManager()
                    .addListener('garage:state', listener)
                    .addListener('garage:shelly:connected', listener))
        }
    },
    mounted() {
        this.serverInteractor.connect()
    },
    methods: {

    },
    computed: {
        capitalizedState: function() {
            return this.state.charAt(0).toUpperCase() + this.state.slice(1).toLowerCase();
        }
    }
})