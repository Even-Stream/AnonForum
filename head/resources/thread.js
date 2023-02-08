const plinks = document.querySelectorAll('a.plink');

plinks.forEach(plink => plink.addEventListener('click', AddIdToNewPost, false));

function AddIdToNewPost(e) {
    let id = e.target.innerText;
    document.getElementById('newpost').value += '>>' + id + '\n';
}