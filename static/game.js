'use strict';

var GameController = function(board, passBtn, black, white) {
    var failedAttempts = 0, timer, turn;
    var title = black + ' vs. ' + white + ' - Go';
    var cells = board.getElementsByClassName('cell');
    for (var i=0; i<cells.length; i++) {
        cells[i].addEventListener('click', clickHandler);
    }
    var color = sessionStorage.getItem('color');
    if (color) {
        document.getElementById('color_'+color).checked = true;
        board.classList.remove('disabled');
    }
    document.forms[0].color[0].addEventListener('change', setColor);
    document.forms[0].color[1].addEventListener('change', setColor);
    passBtn.addEventListener('click', pass);
    connect();

    function connect() {
        var proto = document.location.protocol == 'https:' ? 'wss://' : 'ws://';
        var wsurl = proto + window.location.host + '/live' + window.location.pathname;
        var socket = new WebSocket(wsurl);
        socket.onmessage = messageHandler;
        socket.onclose = closeHandler;
        socket.onopen = function() {
            clearInterval(timer);
            document.title = black + ' vs. ' + white + ' - Go';
            failedAttempts = 0;
            if (color) board.classList.remove('disabled');
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
                switch (g.Board[y][x]) {
                    case 1: cell.classList.add('black'); break;
                    case 2: cell.classList.add('white'); break;
                    default: cell.classList.remove('black', 'white');
                }
            }
        }
        if (2 - g.Turn % 2 == color) {
            var flashTimer = setInterval(function() {
                if (document.title == title) document.title = 'Your Turn - ' + title;
                else document.title = title;
            }, 1000);
            window.addEventListener('focus', function() {
                clearInterval(flashTimer);
                document.title = title;
            });
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

    function setColor(event) {
        color = this.value
        sessionStorage.setItem('color', color);
        board.classList.remove('disabled');
    }

    function clickHandler(event) {
        if (board.classList.contains('disabled')) return;

        if (!color) return;

        var x = indexOf(this);
        var y = indexOf(this.parentNode);
        var url = window.location.href;
        var data = 'color=' + color + '&x=' + x + '&y=' + y;
        ajax('PATCH', url, data, response);
    }

    function pass(event) {
        if (!color) return;

        var data = 'color=' + color + '&pass=true';
        ajax('PATCH', window.location.href, data, response);
    }

    function response() {
        if (this.status >= 300) alert(this.response);
    }
}
