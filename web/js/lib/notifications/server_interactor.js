class ServerInteractor {
    constructor(url, notificationManager, context) {
        this._context = context;
        this._notificationManager = notificationManager;
        this._url = url;
        this._ws = null;
        this._isConnected = false;

        this._notificationManager.notifyListeners({
            key: "connected",
            value: false
        }, this._context);
    }

    get notificationManager() { return this._notificationManager; }

    onMessage(event) {
        this._notificationManager.notifyListeners(JSON.parse(event.data), this._context);
    }

    onOpen(event) {
        this._isConnected = true;

        this._notificationManager.notifyListeners({
            key: "connected",
            value: true
        }, this._context);

        this.sendCommand('get');
    }

    onClose(event) {
        this._isConnected = false;

        this._notificationManager.notifyListeners({
            key: "connected",
            value: false
        }, this._context);

        this.connect();
    }

    connect() {
        this._ws = new WebSocket(this._url);
        
        this._ws.addEventListener('open', this.onOpen.bind(this));
        this._ws.addEventListener('message', this.onMessage.bind(this));
        this._ws.addEventListener('close', this.onClose.bind(this));
    }

    sendCommand(command) {
        if (!this._isConnected) {
            return;
        }
        
        this._ws.send(JSON.stringify({
            cmd: command
        }));
    }

    set(key, value) {
        if (!this._isConnected) {
            return;
        }

        this._ws.send(JSON.stringify({
            cmd: "set",
            key: key,
            value: value
        }));
    }
}