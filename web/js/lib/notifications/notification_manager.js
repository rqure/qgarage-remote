class NotificationManager {
    constructor() {
        this._topics = {};
    }

    addListener(topic, listener) {
        if ( !(topic in this._topics) ) {
            this._topics[topic] = [];
        }

        this._topics[topic].push(listener);

        return this;
    }

    notifyListeners(data, context) {
        if (data.key in this._topics) {
            this._topics[data.key].forEach(listener => {
                listener.onNotification(data.key, data, context);
            });
        }
    }

    get topics() {
        return Object.keys(this._topics);
    }
};