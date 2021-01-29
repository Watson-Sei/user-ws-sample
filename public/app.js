const IAM = {
    token: null,   // トークン
    name: null,    // 名前
    is_join: false // 入室中
};

const Member = {
    0: "マスター"
};

const socket = new WebSocket("ws://localhost:8080/ws")

// STEP1. サーバーへ接続
socket.onopen = function () {
    console.log("接続しました")
    $("#nowconnecting").style.display = "none";
    $("#inputmyname").style.display = "block";
    $("#txt-myname").focus();
};

// STEP2. 名前入力
$("#frm-myname").addEventListener("submit", (e)=>{
    // 規定の送信処理をキャンセル(画面遷移しないなど)
    e.preventDefault();

    const myname = $("#txt-myname");
    if( myname.value === "" ){
        return(false);
    }

    $("#myname").innerHTML = myname.value;
    IAM.name = myname.value;

    socket.send(JSON.stringify({
        event: "join",
        token: IAM.token,
        name: IAM.name
    }))

    // ボタンを無効にする
    $("#frm-myname button").setAttribute("disabled", "disabled");
})

socket.onmessage = function (event) {
    const json = JSON.parse(event.data)
    if(json.event === "token") {
        IAM.token = json.token;
    } else if (json.event === "member-post") {
        const is_me = (json.token === IAM.token);
        addMessage(json, is_me);
    } else if (json.event === "join-result") {
        // 正常に入室できた
        if( json.status ) {
            // 入室フラグを立てる
           IAM.is_join = true;

           // すでにログイン中のメンバー一覧を反映
            if (json.list !== null) {
                for(let i=0; i<json.list.length; i++) {
                    const cur = json.list[i];
                    if( ! (cur.token in Member) ){
                        addMemberList(cur.token, cur.name);
                    }
                }
            }

            // 表示を切り替える
            $("#inputmyname").style.display = "none";
            $("#chat").style.display = "block";
            $("#msg").focus();
        }
        // できなかった場合
        else {
            alert("入室できませんでした");
        }

        // ボタンを有効ん戻す
        $("#frm-myname button").removeAttribute("disabled");
    } else if (json.event === "member-join") {
        if( IAM.is_join ) {
            addMessageFromMaster(`${json.name}さんが入室しました`);
            addMemberList(json.token, json.name)
        }
    } else if (json.event === "quit-result") {
        if(json.status) {
            gotoSTEP1();
        } else {
            alert("退室できませんでした");
        }

        // ボタンを有効に戻す
        $("#frm-quit button").removeAttribute("disabled");
    } else if (json.event === "member-quit") {
        if( IAM.is_join ) {
            const name = Member[json.token];
            addMessageFromMaster(`${name}さんが退室しました`);
            removeMemberList(json.token);
        }
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

$("#frm-quit").addEventListener("submit", (e)=>{
    e.preventDefault();

    if( confirm("本当に退室しますか？") ){
        socket.send(JSON.stringify({
            event: "quit",
            token: IAM.token
        }));

        // ボタンを無効にする
        $("#frm-quit button").setAttribute("disabled", "disabled")
    }
})

socket.onclose = function (event) {
    const list = $("#msglist");
    const li = document.createElement("li");

    li.innerHTML = `<span class="msg-master"><span class="name">master</span>> Connection closed.</span>`

    list.insertBefore(li, list.firstChild);
}

function addMessageFromMaster(msg) {
    addMessage({token: 0, text: msg})
}

function addMessage(msg, is_me=false) {
    const list = $("#msglist");
    const li = document.createElement("li");
    const name = Member[msg.token]

    // マスターの発言
    if (msg.token === 0) {
        li.innerHTML = `<span class="msg-master"><span class="name">${name}</span>> ${msg.text}</span>`;
    } else if ( is_me ) {
        li.innerHTML = `<span class="msg-me"><span class="name">${msg.name}</span>> ${msg.message}</span>`
    } else {
        li.innerHTML = `<span class="msg-member"><span class="name">${msg.name}</span>> ${msg.message}</span>`
    }

    list.insertBefore(li, list.firstChild);
}

function addMemberList(token, name) {
    const list = $("#memberlist");
    const li = document.createElement("li");
    li.setAttribute("id", `member-${token}`);
    if( token == IAM.token ) {
        li.innerHTML = `<span class="member-me">${name}</span>`;
    } else {
        li.innerHTML = name;
    }

    // リストの最後に追加
    list.appendChild(li);

    // 内部変数に保存
    Member[token] = name;
}

function removeMemberList(token) {
    const id = `#member-${token}`;
    if( $(id) !== null ){
        $(id).parentNode.removeChild( $(id) );
    }

    // 内部変数から削除
    delete Member[token];

    console.log(Member)
}

function gotoSTEP1() {
    // NowLoadingから開始
    $("#nowconnecting").style.display = "block";  // NowLoadingを表示
    $("#inputmyname").style.display = "none";     // 名前入力を非表示
    $("#chat").style.display = "none";            // チャットを非表示

    // 自分の情報を初期化
    IAM.token = null;
    IAM.name  = null;
    IAM.is_join = false;

    for( let key in Member ){
        if( key !== "0" ){
            delete Member[key];
        }
    }

    // チャット内容を全て消す
    $("#txt-myname").value = "";     // 名前入力欄 STEP2
    $("#myname").innerHTML = "";     // 名前表示欄 STEP3
    $("#msg").value = "";            // 発言入力欄 STEP3
    $("#msglist").innerHTML = "";    // 発言リスト STEP3
    $("#memberlist").innerHTML = ""; // メンバーリスト STEP3
}