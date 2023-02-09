const selectedcolor = getComputedStyle(document.body).getPropertyValue('--sel-color');

const menubuttons = document.querySelectorAll('td.hop');

menubuttons.forEach(menubutton => menubutton.addEventListener('click', SwitchToRow, false));

function SwitchToRow(e) {
    let visibleid = this.id + 'row';
    menubuttons.forEach(menubutton => {
        if(menubutton.id === this.id) {
            menubutton.style.backgroundColor = selectedcolor;
        } else {menubutton.style.backgroundColor = 'transparent';}

        let currentid = menubutton.id + 'row';
        let menurow = document.getElementById(currentid);

        if(menurow.id !== visibleid) {
            menurow.style.visibility = 'collapse';
        } else {menurow.style.visibility = 'visible';}
    });
}