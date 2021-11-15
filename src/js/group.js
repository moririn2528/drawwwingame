(function () {
    const group_number = 20;

    window.onload = function() {
        let sel = document.getElementById('div_select_group')
        sel.innerHTML="<select name=\"group\" id=\"select_group\">\n"
        for(let i=0;i<group_number;i++){
            sel.innerHTML+="<option value=\""+String(i)+"\">グループ "+String(i+1)+"</option>\n"
        }
        sel.innerHTML+="</select>\n"
    };
    
    async function set_group_func(e){
        const uuid = sessionStorage.getItem("uuid")
        const tempid = sessionStorage.getItem("tempid")
        const group_id = document.getElementById('select_group').value
        const response = await postData("/group",{
            "uuid": uuid,
            "tempid": tempid,
            "group_id": group_id,
        })
        window.location.href="../html/main.html"
    }

    select_button.addEventListener("click",e=>{
        set_group_func(e).catch(err=>{
            console.log("Error: select_button, click, ",err);
        })
    });
}());