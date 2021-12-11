let role_str=["answer","writer"];
let my_role;
let turn=1,all_turn;
let game_in_progress=false;
let turn_minutes;
let theme="";

function isAnswer(){
    return my_role==role_str[0];
}
function isWriter(){
    return my_role==role_str[1];
}

function displayInfo(){
    let elem=document.getElementById("my_role")
    switch(my_role){
        case "answer":
            elem.textContent="回答者";
            break;
        case "writer":
            elem.textContent="書き手";
            break;
        case "viewer":
            elem.textContent="閲覧者";
            break;
        case undefined:
            elem.textContent="";
            break;
        default:
            console.assert(false,"my_role error",my_role);
    }
    elem=document.getElementById("turn_times");
    elem.textContent=String(turn)+"/"+String(all_turn);
    elem=document.getElementById("theme");
    if(theme!=""){
        elem.textContent="お題: "+theme;
    }
}

function displayTimeLeft(seconds){
    let elem=document.getElementById("time_left");
    if(seconds<0){
        elem.textContent="";
        return;
    }
    const tim=String(Math.floor(seconds/60))+":"+("00"+String(seconds%60)).slice(-2);
    elem.textContent=tim;
}

function processGameInitMessage(message){
    const words=message.split(":");
    console.assert(words.length==3, "message error", message);
    all_turn=parseInt(words[0],10);
    turn_minutes=parseInt(words[1],10);
    theme="";
}

function processGameNowMessage(message){
    const words=message.split(":");
    console.assert(words.length==4, "message error", message);
    turn=parseInt(words[0],10);
    game_in_progress=Boolean(words[1]=="true");
    if(words[2]==""){
        displayTimeLeft(-1);
    }else{
        const tim=parseInt(words[2],10);
        displayTimeLeft(tim);
    }
    theme=words[3];
}

function processGameEndMessage(message){
    const words=message.split(":");
    console.assert(words.length==3, "message error", message);
    let mes=[]
    switch(words[0]){
    case "win":
        mes.push("勝ち");
        mes.push("回答者: "+words[2]);
        break;
    case "lose":
        mes.push("負け");
        break;
    default:
        console.assert(false,"message error",message);
    }
    mes.push("お題: "+words[1]);
    addInfoChatMessage(mes.join(", "));
}

function processGameFinishMessage(){
    let ok=window.confirm("ゲーム終了。待機画面に戻ります。");
    if(!ok){
        return;
    }
    window.location.href="../html/role.html";
}

function processGameMessage(message_json){
    const info=message_json.message_info.slice(5);
    const mes=message_json.message;
    switch(info){
    case "role":
        my_role=mes;
        break;
    case "init":
        processGameInitMessage(mes);
        break;
    case "now":
        processGameNowMessage(mes);
        break;
    case "start":
        theme=mes;
        break;
    case "end":
        processGameEndMessage(mes);
        break;
    case "finish":
        processGameFinishMessage();
        break;
    case "time":
        displayTimeLeft(parseInt(mes,10));
        break;
    case "turn":
        turn=parseInt(mes);
        break;
    default:
        console.assert(false,message_json);

    }
    displayInfo();
}