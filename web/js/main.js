async function main() {
    const app = Vue.createApp({});
    
    const context = {
        qDatabaseInteractor: new DatabaseInteractor(),
    };

    registerRemoteComponent(app, context);

    app.mount('#desktop');

    context.qDatabaseInteractor.runInBackground(true);

    CURRENT_LOG_LEVEL=LOG_LEVELS.DEBUG;
}