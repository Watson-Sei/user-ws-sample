const IAM = {
    token: null,
    name: null
};

const socket = new WebSocket("ws://localhost:8080/ws")

// STEP1. サーバーへ接続
socket.onopen = function () {
    console.log("接続しました")
    $("#nowconnecting").style.display = "none";
    $("#inputmyname").style.display = "block";
};

// STEP2. 名前入力
$("#frm-myname").addEventListener("submit", (e)=>{
    e.preventDefault();

    const myname = $("#txt-myname");
    if( myname.value === "" ){
        return(false);
    }

    $("#myname").innerHTML = myname.value;
    IAM.name = myname.value;

    // 表示を切り替える
    $("#inputmyname").style.display = "none";   // 名前入力を非表示
    $("#chat").style.display = "block";
})

socket.onmessage = function (event) {
    const json = JSON.parse(event.data)
    if(json.event == "token") {
        IAM.token = json.token;
    } else if (json.event == "member-post") {
        const is_me = (json.token === IAM.token);
        addMessage(json, is_me);
    }
}

$("#frm-post").addEventListener("submit", (e)=> {
    e.preventDefault();

    const msg = $("#msg");
    if( msg.value === "" ){
        return(false);
    }

    socket.send(JSON.stringify({
        event: "post",
        message: msg.value,
        token: IAM.token,
        name: IAM.name
    }));

    msg.value = "";
})

socket.onclose = function (event) {
    const list = $("#msglist");
    const li = document.createElement("li");

    li.innerHTML = `<span class="msg-master"><span class="name">master</span>> Connection closed.</span>`

    list.insertBefore(li, list.firstChild);
}

function addMessage(msg, is_me=false) {
    const list = $("#msglist");
    const li = document.createElement("li");

    if ( is_me ) {
        li.innerHTML = `<span class="msg-me"><span class="name">${msg.name}</span>> ${msg.message}</span>`
    } else {
        li.innerHTML = `<span class="msg-member"><span class="name">${msg.name}</span>> ${msg.message}</span>`
    }

    list.insertBefore(li, list.firstChild);
}