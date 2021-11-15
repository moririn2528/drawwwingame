const sock = new WebSocket("ws://localhost:1213/ws")

function sendWebSocket(type,text){
    let message={
        "uuid": sessionStorage.getItem("uuid"),
        "tempid": sessionStorage.getItem("tempid"),
        "username": sessionStorage.getItem("username"),
        "type": type,
        "message": text,
    }
    sock.send(JSON.stringify(message))
}

(function () {
    const sock = new WebSocket("ws://localhost:1213/ws")

    sock.addEventListener("open", e => {
        console.log("socket opened")
    })
    sock.addEventListener("message", e => {
        const message_json = JSON.parse(e.data);
        const message=message_json.message;
        switch (message_json.type){
        case "text":
            addChatMessage(message);
            break;
        case "lines":
            addLines(message);
            break;
        default:
            console.assert(false,"message type error");
        }
    })
    sock.addEventListener("close", e => {
        console.log("socket closed")
    })
    sock.addEventListener("error", e=>{
        console.log("socket error",e)
    })
}());