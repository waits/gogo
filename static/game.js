'use strict';

var Game = function(cells) {
    for (var i=0; i<cells.length; i++) {
        cells[i].addEventListener('click', clickHandler);
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
        else
            document.location.reload();
    }
}

document.addEventListener('DOMContentLoaded', function() {
    var cells = document.getElementsByClassName('cell');
    new Game(cells);
});
