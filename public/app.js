const IAM = {
    token: null,
    name: null
};

const socket = new WebSocket("ws://localhost:8080/ws")

socket.onopen = function () {
    console.log("接続しました")
}

socket.onmessage = function (event) {
    const json = JSON.parse(event.data)
    if(json.event == "token") {
        IAM.token = json.token
    } else if (json.event == "member-post") {
        addMessage(json.message)
    }
}

$("#frm-post").addEventListener("submit", (e)=> {
    e.preventDefault();

    const msg = $("#msg");
    if( msg.value === "" ){
        return(false);
    }

    socket.send(JSON.stringify({event: "post", message: msg.value}))

    msg.value = "";
})

function addMessage(message) {
    const list = $("#msglist");
    const li = document.createElement("li");

    li.innerHTML = `<span class="msg-member">${message}</span>`

    list.insertBefore(li, list.firstChild);
}