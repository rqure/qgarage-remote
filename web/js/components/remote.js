function registerRemoteComponent(app, context) {
    return app.component("remote", {
// show a dropdown of garage doors
// show the current state of the selected garage door
// show push button to open/close garage door
// push button is locked to prevent accidental click
// user can unlock the push button by clicking on the lock icon
        template: `
<div>
</div>`,

        data() {
            context.qDatabaseInteractor
                .getEventManager()
                .addEventListener(DATABASE_EVENTS.CONNECTED, this.onDatabaseConnected.bind(this))
                .addEventListener(DATABASE_EVENTS.DISCONNECTED, this.onDatabaseDisconnected.bind(this));

            return {
                database: context.qDatabaseInteractor,
                isDatabaseConnected: false,
                garageDoors: [],
                selectedDoor: null,
                state: null,
                locked: true,
            }
        },

        mounted() {
            this.isDatabaseConnected = this.database.isConnected();
        },

        methods: {
            onDatabaseConnected() {
                this.isDatabaseConnected = true;
            },

            onDatabaseDisconnected() {
                this.isDatabaseConnected = false;
            },
        },

        computed: {
            isRemoteEnabled() {
                return this.isDatabaseConnected;
            }
        }
    })
}