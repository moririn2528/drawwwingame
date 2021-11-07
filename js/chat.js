(function () {
    const sock = new WebSocket("ws://localhost:1213/ws")

    sock.addEventListener("open", e => {
        console.log("opened")
    })
    sock.addEventListener("message", e => {
        console.log("message")
        console.log(e)
    })
    sock.addEventListener("close", e => {
        console.log("closed")
    })
    sock.addEventListener("error", e=>{
        console.log(e)
    })

    test.addEventListener("click",e=>{
        var test_mes={
            "uuid": "AAA",
            "type": "text",
            "message": "test message"
        }
        sock.send(JSON.stringify(test_mes))
    })

    chat_button.addEventListener("click",e=>{
        var elem=document.getElementById("chat_screen");
        var message=document.getElementById("chat_send_message");
        var new_elem=document.createElement("div");
        new_elem.className="chat_message";
        new_elem.textContent=message.value;
        message.value="";
        elem.appendChild(new_elem);
    })
}());