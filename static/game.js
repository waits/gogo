'use strict';

var GameController = function(board, black, white) {
    var failedAttempts = 0, timer;
    var cells = board.getElementsByClassName('cell');
    for (var i=0; i<cells.length; i++) {
        cells[i].addEventListener('click', clickHandler);
    }
    var c = sessionStorage.getItem('color');
    if (c) document.getElementById('color_'+c).checked = true;
    document.forms[0].color[0].addEventListener('change', setColor);
    document.forms[0].color[1].addEventListener('change', setColor);
    connect();

    function connect() {
        var proto = document.location.protocol == 'https:' ? 'wss://' : 'ws://';
        var wsurl = proto + window.location.host + '/live' + window.location.pathname;
        var socket = new WebSocket(wsurl);
        socket.onmessage = messageHandler;
        socket.onclose = closeHandler;
        socket.onerror = errorHandler;
        socket.onopen = function() {
            clearInterval(timer);
            document.title = black + ' vs. ' + white + ' - Go';
            failedAttempts = 0;
            board.classList.remove('disabled');
            console.info('WebSocket connected');
        };
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
        if (failedAttempts == 0) {
            document.title = 'Reconnecting';
            timer = setInterval(function() {
                if (document.title.length < 15) document.title += '.';
                else document.title = 'Reconnecting';
            }, 1000);
        }
        var wait = Math.round(Math.pow(failedAttempts++, 1.5) + 1);
        setTimeout(connect, wait * 1000);
        board.classList.add('disabled');
        console.warn('WebSocket closed, attempt ' + failedAttempts + ', reconnecting in ' + wait + 's');
    }

    function errorHandler(event) {
        alert('There was an error connecting to the server. Please refresh the page.');
        console.error('WebSocket error', event);
    }

    function setColor(event) {
        sessionStorage.setItem('color', this.value);
    }

    function clickHandler(event) {
        if (board.classList.contains('disabled')) return;

        var color = sessionStorage.getItem('color');
        if (!color) {
            alert('You have to select a color.');
            return;
        }

        var x = indexOf(this);
        var y = indexOf(this.parentNode);
        var url = window.location.href;
        var data = 'color=' + color + '&x=' + x + '&y=' + y;
        ajax('PATCH', url, data, requestCallback);
    }

    function requestCallback() {
        if (this.status >= 300) alert(this.response);
    }
}
