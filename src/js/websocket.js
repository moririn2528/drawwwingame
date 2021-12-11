const sock = new WebSocket("ws://localhost:1213/ws")
const DEBUG_MODE=true

function sendWebSocket(type,info,text){
    let message={
        "uuid": sessionStorage.getItem("uuid"),
        "tempid": sessionStorage.getItem("tempid"),
        "username": sessionStorage.getItem("username"),
        "group_id": sessionStorage.getItem("group_id"),
        "type": type,
        "message_info": info,
        "message": text,
    }
    if(typeof DEBUG_MODE == "boolean" && DEBUG_MODE){
        console.log("send message,",message);
    }
    sock.send(JSON.stringify(message))
}

(function () {
    function processInfo(message_json){
        const info=message_json.message_info;
        const mes=message_json.message;
        if(info.startsWith("game")){
            if(typeof processGameMessage != "function"){
                window.location.href="../html/main.html"
                return;
            }
            processGameMessage(message_json);
            return;
        }
        if(mes=="join"){
            if(typeof addMember == "function"){
                addMember(message_json.username);
            }
            return;
        }
        if(mes=="leave"){
            if(typeof removeMember == "function"){
                removeMember(message_json.username);
            }
            return;
        }
        if(mes=="ready"){
            if(typeof changeMemberState == "function"){
                changeMemberState(message_json.username,message_json.message);
            }
            return;
        }
        console.assert(false,message_json)
    }

    sock.addEventListener("open", e => {
        console.log("socket opened");
        sendWebSocket("info","","uuid");
        if(typeof initWebSocketLocal=="function"){
            initWebSocketLocal();
        }
    })
    sock.addEventListener("message", e => {
        const message_json = JSON.parse(e.data);
        const message=message_json.message;
        const message_info=message_json.message_info;
        if(typeof DEBUG_MODE == "boolean" && DEBUG_MODE){
            console.log("recieve message,",message_json);
        }
        const t=message_json.type;
        if(t.startsWith("text")){
            addChatMessage(message_json);
            return;
        }
        if(t.startsWith("mark")){
            drawChatMessageMark(message_json);
            return;
        }
        if(t=="lines"){
            addLines(message);
            return;
        }
        if(t=="info"){
            processInfo(message_json);
            return;
        }
        console.assert(false,"message type error");
    })
    sock.addEventListener("close", e => {
        console.log("socket closed")
        window.alert("接続が切れました。リロードしてください。")
    })
    sock.addEventListener("error", e=>{
        console.log("socket error",e)
    })
}());