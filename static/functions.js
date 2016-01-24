'use strict';

function indexOf(el) {
    return [].slice.call(el.parentNode.children).indexOf(el);
}

function typeOf(object) {
	return Object.prototype.toString.call(object).slice(8, -1);
}

function ajax(method, url, data, callback) {
    var request = new XMLHttpRequest();
    request.open(method, url, true);
    request.onload = function() {
        if (callback) callback.call(this);
    };

    if (method === 'GET') {
        return request.send();
    }
    else if (typeOf(data) === 'Object') {
        request.setRequestHeader('Content-Type', 'application/json');
        request.send(JSON.stringify(data));
    }
    else {
    	request.setRequestHeader('Content-Type', 'application/x-www-form-urlencoded');
        request.send(data);
    }
}
