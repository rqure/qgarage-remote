async function main() {
    const app = Vue.createApp({});
    
    const context = {
        qEntityStore: QEntityStore({
            port: ":20000",
        }),
    };

    registerRemoteComponent(app, context);

    app.mount('#desktop');

    qEntityStore.runInBackground(true);

    CURRENT_LOG_LEVEL=LOG_LEVELS.DEBUG;
}