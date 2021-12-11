(function () {

    function change_role(e){
        const uuid = sessionStorage.getItem("uuid")
        const tempid = sessionStorage.getItem("tempid")
        const ans = document.getElementById("can_answer").checked
        const writ = document.getElementById("can_writer").checked
        res = postData("/group/role",{
            "uuid":uuid,
            "tempid":tempid,
            "admin":false,
            "can_answer":ans,
            "can_writer":writ,
        })
    }

    can_answer.addEventListener("change",change_role)
    can_writer.addEventListener("change",change_role)

    is_ready.addEventListener("click",e=>{
        sendWebSocket("info","","ready")
    })
}());

let usernames=[];
function addMember(username){
    for(let i=0;i<usernames.length;i++){
        if(usernames[i]==username)return;
    }
    const id=usernames.length;
    let elem=document.getElementById("stay_member");
    let name_elem=document.createElement("div");
    name_elem.textContent=username;
    elem.appendChild(name_elem);
    usernames.push(username);
}

function removeMember(username){
    let elem=document.getElementById("stay_member");
    for(let i=0;i<usernames.length;i++){
        if(usernames[i]!=username)continue;
        usernames.splice(i,1);
        elem.removeChild(elem.childNodes[i]);
        return;
    }
    console.assert(false,"removeMember error, ",usernames,username);
}
    
function changeMemberState(username,mes){
    let elem=document.getElementById("stay_member");
    if(mes=="ready"){
        for(let i=0;i<usernames.length;i++){
            if(usernames[i]!=username)continue;
            elem.childNodes[i].classList.add("is_ready");
            return;
        }
        console.assert(false,username,mes);
        return;
    }
    console.assert(false,username,mes);
}

function initWebSocketLocal(){
    sendWebSocket("info","","join");
}