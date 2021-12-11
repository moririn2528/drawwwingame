
var canvas_space = canvas_space || {};
(function (global) {
    const canvas_background_color="#f5f5f5";
    let _=canvas_space;
    let canvas=document.getElementById('canv');
    let cv=canvas.getContext('2d');
    cv.beginPath();
    cv.fillStyle = canvas_background_color;
    cv.fillRect(0, 0, 700, 400);
    cv.lineJoin="round";
    let brush_size=3;
    let brush_color="#000000";
    let brush_alpha=1.0;

    let bef_x=""
    let bef_y=""

    let send_pos_x=[], send_pos_y=[];//int array

    function drawLine(x1,y1,x2,y2,size,color,alpha){
        if(x1==x2 && y1==y2)return;
        cv.beginPath();
        cv.moveTo(x1,y1);
        cv.lineTo(x2,y2);
        cv.lineCap="round";
        cv.lineWidth=size;
        cv.strokeStyle=color;
        cv.stroke();
    }
    
    _.drawLines = function(xs,ys,size,color,alpha){
        console.assert(xs.length==ys.length);
        if(xs.length<=1)return;
        cv.beginPath();
        cv.moveTo(xs[0],ys[0]);
        for(let i=1;i<xs.length;i++){
            cv.lineTo(xs[i],ys[i]);
        }
        cv.lineCap="round";
        cv.lineWidth=size;
        cv.strokeStyle=color;
        cv.stroke();
    };

    function drawMe(x, y){
        if(!isWriter())return;
        x=~~x,y=~~y;
        if(x<0 || y<0)return;
        send_pos_x.push(x),send_pos_y.push(y);
        if(bef_x==""){
            bef_x=x,bef_y=y;
            return;
        }
        drawLine(x,y,bef_x,bef_y,brush_size,brush_color,brush_alpha);
        bef_x=x,bef_y=y;
    }

    _.sendLines=function(){
        if(send_pos_x.length<=1)return;
        let poss=[];
        console.assert(send_pos_x.length==send_pos_y.length);
        for(let i=0;i<send_pos_x.length;i++){
            poss.push(send_pos_x[i].toString()+":"+send_pos_y[i].toString());
        }
        message=[
            brush_size.toString(),brush_color,brush_alpha.toString(),
            poss.join("-")].join(",");
        sendWebSocket("lines","",message);
        if(bef_x=="")send_pos_x=[],send_pos_y=[];
        else send_pos_x=[bef_x],send_pos_y=[bef_y];
    }

    canvas.addEventListener('mousemove', e => {
        if(!isWriter())return;
        if(e.buttons&1){
            let rect=e.target.getBoundingClientRect();
            let x=e.clientX-rect.left;
            let y=e.clientY-rect.top;
            drawMe(x,y);
        }
    });

    canvas.addEventListener('mousedown', e => {
        if(!isWriter())return;
        if(e.button==0){
            let rect=e.target.getBoundingClientRect();
            let x=e.clientX-rect.left;
            let y=e.clientY-rect.top;
            drawMe(x,y);
        }
    });

    canvas.addEventListener('mouseup', drawEnd);
    canvas.addEventListener('mouseout', e => {
        if(e.buttons&1){
            drawEnd(e);
        }
    });

    let canv_color=document.getElementById("canv_color");
    canv_color.addEventListener("change",e=>{
        brush_color=canv_color.value;
    });
    canv_color.addEventListener("click",e=>{
        brush_color=canv_color.value;
    });
    function drawEnd(e){
        if(!isWriter())return;
        let rect=e.target.getBoundingClientRect();
        let x=e.clientX-rect.left;
        let y=e.clientY-rect.top;
        drawMe(x,y);
        bef_x="",bef_y="";
        _.sendLines();
    }
    
    let canv_brush_size=document.getElementById("canv_brush_size");
    canv_brush_size.addEventListener("change",e=>{
        brush_size=parseInt(canv_brush_size.value,10);
    });

    
    _.canvasClear=function(){
        cv.fillStyle = canvas_background_color;
        cv.fillRect(0, 0, 700, 400);
    }
    
    document.getElementById("canvas_brush_eraser").addEventListener("click",e=>{
        brush_color=canvas_background_color;
    });
    document.getElementById("canvas_clear").addEventListener("click",e=>{
        _.canvasClear();
        sendWebSocket("lines","","#clear");
    });

    setInterval("canvas_space.sendLines()",50);
}(this));

function addLines(text){
    if(text=="#clear"){
        canvas_space.canvasClear();
        return;
    }
    const setting=text.split(",");
    console.assert(setting.length==4);
    const size=parseInt(setting[0],10);
    const color=setting[1];
    const alpha=parseInt(setting[2],10);
    const poss=setting[3].split("-");
    let xs=[],ys=[];
    for(let i=0;i<poss.length;i++){
        const pos=poss[i].split(":");
        console.assert(pos.length==2);
        const x=parseInt(pos[0],10);
        const y=parseInt(pos[1],10);
        xs.push(x);
        ys.push(y);
    }
    canvas_space.drawLines(xs,ys,size,color,alpha);
}
