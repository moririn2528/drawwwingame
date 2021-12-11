let chat_tab_index=0;

jQuery(function($){
    function sendMessage(){
        const message=$("#chat_send_message").val();
        addMyChatMessage(message);
        $("#chat_send_message").val("");
    }

    $("#chat_button").click(sendMessage);
    $("#chat_send_message").keydown(function(e){
        if(e.ctrlKey && e.code == "Enter"){
            sendMessage();
        }
    });

    $(".chat_tab").click(function(){
        $(".chat_tab.is_active").removeClass("is_active");
        $(".chat_screen.is_active").removeClass("is_active");
        $(this).addClass("is_active");
        const index = $(this).index();
        chat_tab_index=index;
        $(".chat_screen").eq(index).addClass("is_active");
    });

});

function sendMessageMark(id, cnt){
    if(!isWriter()){
        return;
    }
    const marks_str="ABC";
    const chat_type=role_str[chat_tab_index];
    sendWebSocket("mark:"+chat_type,String(id),marks_str[cnt]);
}

let my_message_count=0;
let my_messages_id=[];
function addChatMessageDiv(elem,username,message){
    let msg_comp=document.createElement("div");
    msg_comp.className="chat_message_text";
    let name=document.createElement("div");
    name.className="chat_message_username";
    name.textContent=username;
    msg_comp.appendChild(name);
    let msg=document.createElement("div");
    msg.className="chat_message_content_text";
    msg.textContent=message;
    msg_comp.appendChild(msg);
    elem.appendChild(msg_comp);
}

function addMyChatMessage(message){
    message=message.trim();
    const chat_type=role_str[chat_tab_index];
    if(message=="")return;
    if(chat_type!=my_role){
        window.alert("このチャットには書き込めません。");
        return;
    }
    my_id=my_message_count;
    my_message_count++;
    my_messages_id.push(my_id);
    
    let elem=document.getElementById("chat_sending_message_"+chat_type);
    addChatMessageDiv(elem,sessionStorage.getItem("username"),message)

    sendWebSocket("text:"+chat_type,String(my_id),message);
}

function addInfoChatMessage(message){
    let elem=document.getElementById("chat_sent_message_answer");
    addChatMessageDiv(elem,"info",message);
}

function addChatMessage(message_json){
    const id=parseInt(message_json.id,10);
    const message_info=message_json.message_info;
    const message_type=message_json.type.slice(5);
    const message=message_json.message;
    const username=message_json.username;
    if(message_info!="#before"){
        const words=message_info.split(":");
        console.assert(2<=words.length,"message info error",message_json);
        const uuid_str=words[0];
        const my_id=parseInt(words[1],10);
    
        //erase chat_sending_text
        let elem=document.getElementById("chat_sending_message_"+message_type);
        if(uuid_str==sessionStorage.getItem("uuid")){
            let flag=false;
            for(let i=0;i<my_messages_id.length;i++){
                if(my_messages_id[i]!=my_id)continue;
                my_messages_id.splice(i,1);
                const msg=elem.childNodes[i];
                elem.removeChild(msg);
                flag=true;
            }
            console.assert(flag,"no exist chat_sending_text",my_messages_id,message_json);
        }
    }

    elem=document.getElementById("chat_sent_message_"+message_type);
    let msg=document.createElement("div");
    msg.className="chat_message";
    addChatMessageDiv(msg,username,message);

    let msg_marks_perc=document.createElement("div");
    msg_marks_perc.className="chat_message_marks_and_percentage";

    let msg_marks=document.createElement("div");
    msg_marks.className="chat_message_marks";
    let perc=document.createElement("div");
    perc.className="chat_message_mark_percentage";
    const marks_str="ABC";
    for(let i=0;i<3;i++){
        let but=document.createElement("div");
        but.className="chat_message_mark chat_message_mark_"+marks_str[i];
        if(isWriter()){
            but.addEventListener("click", e=>sendMessageMark(id,i));
        }
        msg_marks.appendChild(but);

        let p=document.createElement("div");
        p.id="chat_message_percentage_"+String(id)+"_"+String(i);
        p.className="chat_message_mark_percentage_in chat_message_mark_percentage_"+marks_str[i];
        perc.appendChild(p);
    }
    msg_marks_perc.appendChild(msg_marks);
    msg_marks_perc.appendChild(perc);
    msg.appendChild(msg_marks_perc);
    elem.appendChild(msg);
}

function drawChatMessageMark(message_json){
    const id=parseInt(message_json.id,10);
    const message=message_json.message;
    const counts=message.split(":");
    console.assert(4<=counts.length,"counts length error",message_json);
    const usercnt=parseInt(counts[3],10);
    for(let i=0;i<3;i++){
        const cnt=parseInt(counts[i],10);
        let elem=document.getElementById("chat_message_percentage_"+String(id)+"_"+String(i));
        elem.style.width=String(Math.round(100*cnt/usercnt))+"%";
    }
}