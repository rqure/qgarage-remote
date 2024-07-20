function registerRemoteComponent(app, context) {
    return app.component("remote", {
        // show a dropdown of garage doors
        // show the current state of the selected garage door
        // show push button to open/close garage door
        // push button is locked to prevent accidental click
        // user can unlock the push button by clicking on the lock switch
        template: `
<div class="container mt-5">
    <div class="card">
        <div class="card-header">
            <h3>Garage Door Remote</h3>
        </div>
        <div class="card-body">
            <div class="row">
                <div class="col-auto">
                    <select class="form-control" id="garageDoorSelect" v-model="selectedGarageDoor">
                        <option v-for="door in garageDoors" :key="door.getId()" :value="door.getId()">
                            {{ door.getName() }}
                        </option>
                    </select>
                </div>
                <div class="col-auto">
                    <button class="btn btn-primary" :disabled="isButtonLocked" @click="toggleGarageDoor">
                        Open / Close
                    </button>
                </div>
                <div class="col-auto">
                    <div class="form-check form-switch mt-2">
                        <input type="checkbox" class="form-check-input" id="lockSwitch" v-model="isButtonLocked" />
                        <label class="form-check-label" for="lockSwitch">
                        </label>
                    </div>
                </div>
            </div>
            <div class="garage">
            </div>
        </div>
    </div>
</div>`,
        data() {
            context.qDatabaseInteractor
                .getEventManager()
                .addEventListener(DATABASE_EVENTS.CONNECTED, this.onDatabaseConnected.bind(this))
                .addEventListener(DATABASE_EVENTS.DISCONNECTED, this.onDatabaseDisconnected.bind(this))
                .addEventListener(DATABASE_EVENTS.QUERY_ALL_ENTITIES, this.onQueryAllEntities.bind(this));

            return {
                database: context.qDatabaseInteractor,
                isDatabaseConnected: false,
                garageDoors: [],
                selectedGarageDoor: 1,
                isButtonLocked: true,
                showGarage: false,
                percentClosed: 0,
            };
        },
        computed: {
            
        },
        
        mounted() {
            this.isDatabaseConnected = this.database.isConnected();

            if (this.isDatabaseConnected) {
                this.onDatabaseConnected();
            }
        },

        methods: {
            onDatabaseConnected() {
                this.isDatabaseConnected = true;
                this.database.queryAllEntities("GarageDoor");
            },

            onDatabaseDisconnected() {
                this.isDatabaseConnected = false;
            },

            onQueryAllEntities(entities) {
                this.garageDoors = entities;
            },
        },
    })
}