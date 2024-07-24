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
            <div class="row">
                <div class="col"></div>
                <div class="col-auto">
                    <select class="form-control" id="garageDoorSelect" v-model="selectedGarageDoorId">
                        <option v-for="door in garageDoors" :key="door.getId()" :value="door.getId()">
                            {{ door.getName() }}
                        </option>
                    </select>
                </div>
                <div class="col-auto">
                    <div class="form-check form-switch mt-2">
                        <input type="checkbox" class="form-check-input" id="lockSwitch" v-model="isButtonLocked" />
                        <label class="form-check-label" for="lockSwitch">
                        </label>
                    </div>
                </div>
                <div class="col"></div>
            </div>
        </div>
        <div class="card-body">
            <div class="garage">
                <button type="button" class="btn btn-outline-light btn-lg garage-inner" :disabled="isButtonLocked" @click="onDoorButtonPressed">
                    {{nextGarageStatus}} <div v-if="isButtonLocked" style="display:inline;">(<i class="fa fa-lock"></i>)</div>
                </button>                
                <button type="button" class="garage-inner-fill btn btn-secondary btn-lg" :style="garageStyle" :disabled="true">
                </button>
            </div>
        </div>
    </div>
</div>`,
        data() {
            context.qDatabaseInteractor
                .getEventManager()
                .addEventListener(DATABASE_EVENTS.CONNECTED, this.onDatabaseConnected.bind(this))
                .addEventListener(DATABASE_EVENTS.DISCONNECTED, this.onDatabaseDisconnected.bind(this))
                .addEventListener(DATABASE_EVENTS.REGISTER_NOTIFICATION_RESPONSE, this.onRegisterNotification.bind(this))
                .addEventListener(DATABASE_EVENTS.NOTIFICATION, this.onNotification.bind(this))
                .addEventListener(DATABASE_EVENTS.READ_RESULT, this.onReadResult.bind(this))
                .addEventListener(DATABASE_EVENTS.QUERY_ALL_ENTITIES, this.onQueryAllEntities.bind(this));

            return {
                database: context.qDatabaseInteractor,
                isDatabaseConnected: false,
                garageDoors: [],
                selectedGarageDoorId: "",
                isButtonLocked: true,
                notificationTokens: [],
                percentClosed: 0,
                lastGarageStatus: ""
            };
        },

        computed: {
            garageStatus() {
                if( this.percentClosed === 100 ) {
                    return "Closed"
                } else if ( this.percentClosed === 0 ) {
                    return "Opened";
                } else if ( this.lastGarageStatus === "Closed" ) {
                    return "Opening";
                } else if ( this.lastGarageStatus === "Opened" ) {
                    return "Closing";
                } else {
                    return "Partially Opened";
                }
            },

            nextGarageStatus() {
                if ( this.percentClosed === 100 ) {
                    return "Open";
                } else if ( this.percentClosed === 0 ) {
                    return "Close";
                } else if ( this.lastGarageStatus === "Closed" ) {
                    return "Stop Opening";
                } else if ( this.lastGarageStatus === "Opened" ) {
                    return "Stop Closing";
                } else {
                    return "Open / Close";
                }
            },

            doorButtonDisabled() {
                return this.isButtonLocked || !this.isDatabaseConnected || this.garageDoors.length === 0 || !this.selectedGarageDoorId || !this.garageDoors[this.selectedGarageDoorId];
            },

            garageStyle() {
                return {
                    height: this.percentClosed + "%",
                    transition: "height 0.5s linear"
                }
            }
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

            onDoorButtonPressed() {
                this.isButtonLocked = true;
                
                const value = new proto.qdb.Int();
                value.setRaw(0);
                const valueAsAny = new proto.google.protobuf.Any();
                valueAsAny.pack(value.serializeBinary(), qMessageType(value));

                if( this.percentClosed === 100 ) {
                    this.database.write([{
                        id: this.selectedGarageDoorId,
                        field: "OpenTrigger",
                        value: valueAsAny
                    }]);
                } else if ( this.percentClosed === 0 ) {
                    this.database.write([{
                        id: this.selectedGarageDoorId,
                        field: "CloseTrigger",
                        value: valueAsAny
                    }]);
                } else if ( this.lastGarageStatus === "Closed" ) {
                    this.database.write([{
                        id: this.selectedGarageDoorId,
                        field: "CloseTrigger",
                        value: valueAsAny
                    }]);
                } else if ( this.lastGarageStatus === "Opened" ) {
                    this.database.write([{
                        id: this.selectedGarageDoorId,
                        field: "OpenTrigger",
                        value: valueAsAny
                    }]);
                } else {
                    this.database.write([{
                        id: this.selectedGarageDoorId,
                        field: "CloseTrigger",
                        value: valueAsAny
                    }]);
                }
            },

            onQueryAllEntities(result) {
                this.garageDoors = result["entities"];
                this.selectedGarageDoorId = this.garageDoors[0].getId();
            },

            onRegisterNotification(event) {
                this.notificationTokens = event.tokens;
            },

            onNotification(event) {
                const notification = event.notification.getCurrent();
                const protoClass = notification.getValue().getTypeName().split('.').reduce((o,i)=> o[i], proto);
                this.percentClosed = protoClass.deserializeBinary(notification.getValue().getValue_asU8()).getRaw();

                if ( this.garageStatus === "Closed" || this.garageStatus === "Opened" ) {
                    this.lastGarageStatus = this.garageStatus;
                }
            },

            onReadResult(event) {
                const protoClass = event[0].getValue().getTypeName().split('.').reduce((o,i)=> o[i], proto);
                this.percentClosed = protoClass.deserializeBinary(event[0].getValue().getValue_asU8()).getRaw();
            }
        },
        
        watch: {
            selectedGarageDoorId: function(newVal) {
                if (this.notificationTokens.length > 0) {
                    this.database.unregisterNotifications(this.notificationTokens.slice());
                    this.notificationTokens = [];
                }

                this.database.registerNotifications([
                    { id: newVal, field: "PercentClosed", notifyOnChange: true }
                ]);

                this.database.read([
                    { id: newVal, field: "PercentClosed" }
                ]);
            },
        }
    })
}