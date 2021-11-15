(function () {
    const sock = new WebSocket("ws://localhost:1213/ws")

    chat_button.addEventListener("click",e=>{
        let message_box=document.getElementById("chat_send_message");
        const message=message_box.value;
        if(message=="")return;
        sendWebSocket("text",message);
        message_box.value=""
    })
}());

function addChatMessage(message){
    let elem=document.getElementById("chat_screen");
    let new_elem=document.createElement("div");
    new_elem.className="chat_message";
    new_elem.textContent=message;
    elem.appendChild(new_elem);
}