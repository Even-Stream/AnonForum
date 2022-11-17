const idmap = new Map();
idmap.set('Main', 'mrow');
idmap.set('News', 'nrow');
idmap.set('About', 'arow');
idmap.set('Rules', 'rrow');
idmap.set('Boards', 'brow');

function showMask(clicked_id) {
    var sel = document.getElementById(clicked_id);
    var crow = document.getElementById(idmap.get(clicked_id));

    if (crow.style.visibility=='collapse') {
        crow.style.visibility = 'visible';
    }

    if (sel.style.backgroundColor == 'transparent') {
        sel.style.backgroundColor = 'rgb(240, 193, 197)';
    }

    idmap.forEach((value, key) => {
        if (key != clicked_id) {
		document.getElementById(key).style.backgroundColor = 'transparent';
		document.getElementById(value).style.visibility = 'collapse';
	}
    });

}