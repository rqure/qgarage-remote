class ServerInteractor {
    constructor(url, notificationManager, context) {
        this._context = context;
        this._notificationManager = notificationManager;
        this._url = url;
        this._ws = null;
        this._isConnected = false;
    }

    get notificationManager() { return this._notificationManager; }

    onMessage(event) {
        this._notificationManager.notifyListeners(JSON.parse(event.data), this._context);
    }

    onOpen(event) {
        this._isConnected = true;

        this.sendCommand('get')
    }

    onClose(event) {
        this._isConnected = false;

        this.connect();
    }

    connect() {
        this._ws = new WebSocket(this._url);
        
        this._ws.addEventListener('open', this.onOpen.bind(this));
        this._ws.addEventListener('message', this.onMessage.bind(this));
        this._ws.addEventListener('close', this.onClose.bind(this));
    }

    sendCommand(command) {
        this._ws.send(JSON.stringify({
            cmd: command
        }));
    }

    set(key, value) {
        this._ws.send(JSON.stringify({
            cmd: "set",
            key: key,
            value: value
        }));
    }
}