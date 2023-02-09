const plinks = document.querySelectorAll('a.plink');

plinks.forEach(plink => plink.addEventListener('click', AddIdToNewPost, false));

function AddIdToNewPost(e) {
    let id = e.target.innerText;
    document.getElementById('newpost').value += '>>' + id + '\n';
}


const replies = document.querySelectorAll('a.preview');

replies.forEach(reply => 
    reply.addEventListener('mouseover', GetPostData, 
    {once : true, capture: false})
);

async function GetPostData (e) {
    let url = e.target.getAttribute('prev-get');
    let response = await fetch(url)
    response.text().then((text) => {
        e.target.insertAdjacentHTML('afterend', 
            '<box class="prev">' + text + '</box>')
    });
}