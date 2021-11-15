(function () {

    async function login_func(e){
        const username=document.getElementById("username").value;
        const password_row=document.getElementById("password").value;
        const password= await createHashPassword(password_row)
        document.getElementById("password").value="";
        const response = await postData("/login",{
            "username": username,
            "password": password,
        })
        const res = await response.json();
        sessionStorage.setItem("uuid", res.uuid)
        sessionStorage.setItem("tempid", res.tempid)
        sessionStorage.setItem("username", res.username)
        window.location.href="../html/group.html"
    }

    login.addEventListener("click",e=>{
        login_func(e).catch(err=>{
            console.log(err);
        })
    })
}());