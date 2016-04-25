'use strict';

var GameController = function(board, passBtn, key, black, white) {
    var notice;
    var title = black + ' vs. ' + white + ' - Go';
    var cells = board.getElementsByClassName('cell');
    for (var i=0; i<cells.length; i++) {
        cells[i].addEventListener('click', clickHandler);
    }
    if (passBtn) {
        var failedAttempts = 0, timer, turn;
        var color = document.cookie.substr(document.cookie.indexOf(key) + 17, 5);
        board.classList.add(color);
        passBtn.addEventListener('click', pass);
    }
    connect();

    function connect() {
        var proto = document.location.protocol == 'https:' ? 'wss://' : 'ws://';
        var wsurl = proto + window.location.host + '/live' + window.location.pathname.replace('watch', 'game');
        var socket = new WebSocket(wsurl);
        socket.onmessage = messageHandler;
        socket.onclose = closeHandler;
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
        if (g.Last === 'f') document.location.reload();
        else if (black != g.Black ^ white != g.White) document.location.reload();
        else if (!document.getElementById('turn')) return;

        document.getElementById('turn').textContent = g.Turn;
        document.getElementById('blackscr').textContent = g.BlackScr;
        document.getElementById('whitescr').textContent = g.WhiteScr;
        for (var y=0; y<g.Board.length; y++) {
            for (var x=0; x<g.Board[y].length; x++) {
                var cell = cells[y*g.Board.length+x];
                var stone = cell.children[1];
                switch (g.Board[y][x]) {
                    case 1: stone.classList.add('black'); break;
                    case 2: stone.classList.add('white'); break;
                    default:
                        stone.classList.remove('black', 'white');
                        stone.classList.add('hide');
                }
                if (g.Last == x * 19 + y) {
                    stone.classList.add('last');
                } else {
                    stone.classList.remove('last');
                }
            }
        }

        if (color && (g.Turn % 2 == 1) == (color == 'black') && !document.getElementById('turn-notice')) {
            notice = document.createElement('div');
            notice.id = 'turn-notice';
            notice.className = 'notice';
            notice.textContent = 'Your turn!';
            document.body.insertBefore(notice, document.getElementById('title'));
            if (!document.hasFocus()) flashTitle();
        }
    }

    function flashTitle() {
        var flashTimer = setInterval(function() {
            if (document.title == title) document.title = 'Your Turn - ' + title;
            else document.title = title;
        }, 1000);
        window.addEventListener('focus', function() {
            clearInterval(flashTimer);
            document.title = title;
        });
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

    function clickHandler(event) {
        if (board.classList.contains('disabled') || board.classList.contains('inactive')) return;

        var x = indexOf(this);
        var y = indexOf(this.parentNode);
        var url = window.location.href;
        var data = 'color=' + color + '&x=' + x + '&y=' + y;
        ajax('PUT', url, data, response);
    }

    function pass(event) {
        if (board.classList.contains('disabled') || board.classList.contains('inactive')) return;

        var data = 'color=' + color + '&pass=true';
        ajax('PUT', window.location.href, data, response);
    }

    function response() {
        if (this.status >= 300) alert(this.response);
        if (notice) notice.remove();
    }
}
