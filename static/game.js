'use strict';

var Game = function(cells) {
    var failedAttempts = 0;
    for (var i=0; i<cells.length; i++) {
        cells[i].addEventListener('click', clickHandler);
    }
    connect();

    function connect() {
        var proto = document.location.protocol == 'https:' ? 'wss://' : 'ws://'
        var wsurl = proto + window.location.host + '/live' + window.location.pathname;
        var socket = new WebSocket(wsurl);
        socket.onmessage = messageHandler;
        socket.onclose = closeHandler;
        socket.onerror = errorHandler;
        socket.onopen = function() {failedAttempts = 0; console.info('WebSocket connected');};
    }

    function messageHandler(event) {
        var g = JSON.parse(event.data);
        document.getElementById('turn').textContent = g.Turn;
        document.getElementById('blackscr').textContent = g.BlackScr;
        document.getElementById('whitescr').textContent = g.WhiteScr;
        for (var y=0; y<g.Board.length; y++) {
            for (var x=0; x<g.Board[y].length; x++) {
                var cell = cells[y*g.Board.length+x];
                var color = null;
                switch (g.Board[y][x]) {
                    case 1: cell.classList.add('black'); break;
                    case 2: cell.classList.add('white'); break;
                    default: cell.classList.remove('black', 'white');
                }
            }
        }
    }

    function closeHandler(event) {
        var wait = Math.round(Math.pow(failedAttempts++, 1.5) + 1);
        setTimeout(connect, wait * 1000);
        console.warn('WebSocket closed, attempt ' + failedAttempts + ', reconnecting in ' + wait + 's');
    }

    function errorHandler(event) {
        alert('There was an error connecting to the server. Please refresh the page.');
        console.error('WebSocket errored', event);
    }

    function clickHandler(event) {
        var x = indexOf(this);
        var y = indexOf(this.parentNode);
        var url = window.location.href;
        var data = 'x=' + x + '&y=' + y;
        ajax('PATCH', url, data, requestCallback);
    }

    function requestCallback() {
        if (this.status >= 300)
            alert(this.response);
    }
}

document.addEventListener('DOMContentLoaded', function() {
    var cells = document.getElementsByClassName('cell');
    cells.length && new Game(cells);
});
