const idmap = new Map();
idmap.set('Main', 'mrow');
idmap.set('News', 'nrow');
idmap.set('About', 'arow');
idmap.set('Rules', 'rrow');
idmap.set('Boards', 'brow');

var style = getComputedStyle(document.body)

function showMask() {
    var sel = document.getElementById(this.id);
    var crow = document.getElementById(idmap.get(this.id));

    if (crow.style.visibility=='collapse') {
        crow.style.visibility = 'visible';
    }

    if (sel.style.backgroundColor == 'transparent') {
        sel.style.backgroundColor = style.getPropertyValue('--sel-color');
    }

    idmap.forEach((value, key) => {
        if (key != this.id) {
		document.getElementById(key).style.backgroundColor = 'transparent';
		document.getElementById(value).style.visibility = 'collapse';
	}
    });

}

idmap.forEach((value, key) => {
	document.getElementById(key).addEventListener("click", showMask)
});