class ClockGatewayListener extends NotificationListener {
    constructor(vm) {
        super()
        this._vm = vm;
    }

    onNotification(key, data, context) {
        this._vm.dt = new Date(data.value);
        this._vm.time = this._vm.dt.toLocaleTimeString(undefined, {
            hour: 'numeric',
            minute: 'numeric'
        })
        this._vm.date = this._vm.dt.toLocaleDateString(undefined, {
            weekday: 'long',
            day: 'numeric',
            month: 'long'
        })
    }
}

const remoteApp = Vue.createApp({
    data() {
        return {
            time: "",
            date: "",
            dt: "",
            serverInteractor:
                new ServerInteractor(`ws://${window.location.hostname}:20000/ws`, new NotificationManager()
                    .addListener('clock-gateway:datetime', new ClockGatewayListener(this)))
        }
    },
    mounted() {
        this.serverInteractor.connect()
    },
    methods: {

    },
    computed: {
        compute() {

        }
    }
})