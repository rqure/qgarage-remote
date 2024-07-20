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
                        <option v-for="door in garageDoors" :key="door.id" :value="door.id">
                            {{ door.name }}
                        </option>
                    </select>
                </div>
                <div class="col-auto">
                    <span>{{ currentState }}</span>
                </div>
                <div class="col-auto">
                    <button class="btn btn-primary" :disabled="isButtonLocked" @click="toggleGarageDoor">
                        {{ buttonLabel }}
                    </button>
                </div>
                <div class="col-auto">
                    <div class="form-check mt-2">
                        <input type="checkbox" class="form-check-input" id="lockSwitch" v-model="isButtonLocked" />
                        <label class="form-check-label" for="lockSwitch">
                            {{ lockLabel }}
                        </label>
                    </div>
                </div>
            </div>
            <div v-if="showGarage" class="garage-door">
                <transition name="door">
                    <div class="garage-door-close-to-open"></div>
                </transition>
            </div>
            <div v-else class="garage-door">
                <transition name="door">
                    <div class="garage-door-close-to-open"></div>
                </transition>
            </div>
        </div>
    </div>
</div>`,
        data() {
            return {
                garageDoors: [
                    { id: 1, name: "Garage Door 1", state: "Closed" },
                    { id: 2, name: "Garage Door 2", state: "Open" },
                ],
                selectedGarageDoor: 1,
                isButtonLocked: true,
                showGarage: false,
            };
        },
        computed: {
            currentState() {
                const door = this.garageDoors.find(
                    (door) => door.id === this.selectedGarageDoor
                );
                return door ? door.state : "";
            },
            buttonLabel() {
                return this.currentState === "Open" ? "Close Door" : "Open Door";
            },
            lockLabel() {
                return this.isButtonLocked ? "Unlock Button" : "Lock Button";
            },
        },
        watch: {
            currentState(newState) {
                this.showGarage = newState === "Closed";
            },
        },
        methods: {
            toggleGarageDoor() {
                const door = this.garageDoors.find(
                    (door) => door.id === this.selectedGarageDoor
                );
                if (door) {
                    door.state = door.state === "Open" ? "Closed" : "Open";
                }
            },
        },
        // data() {
        //     context.qDatabaseInteractor
        //         .getEventManager()
        //         .addEventListener(DATABASE_EVENTS.CONNECTED, this.onDatabaseConnected.bind(this))
        //         .addEventListener(DATABASE_EVENTS.DISCONNECTED, this.onDatabaseDisconnected.bind(this));

        //     return {
        //         database: context.qDatabaseInteractor,
        //         isDatabaseConnected: false,
        //         garageDoors: [],
        //         selectedDoor: null,
        //         state: null,
        //         locked: true,
        //     }
        // },

        // mounted() {
        //     this.isDatabaseConnected = this.database.isConnected();
        // },

        // methods: {
        //     onDatabaseConnected() {
        //         this.isDatabaseConnected = true;
        //     },

        //     onDatabaseDisconnected() {
        //         this.isDatabaseConnected = false;
        //     },
        // },

        // computed: {
        //     isRemoteEnabled() {
        //         return this.isDatabaseConnected && this.garageDoors.length > 0 && this.selectedDoor !== null && this.state !== null && !this.locked;
        //     }
        // }
    })
}