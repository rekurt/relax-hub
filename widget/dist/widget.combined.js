(function() {
  // Inject CSS
  if (!document.getElementById('bani-widget-styles')) {
    const style = document.createElement('style');
    style.id = 'bani-widget-styles';
    style.textContent = `.bani-widget{font-family:Arial,sans-serif;background:#f5f5f5;border:1pxsolid#ddd;border-radius:8px;padding:20px;max-width:500px;box-shadow:02px8pxrgba(0,0,0,0.1);}.bani-widget*{box-sizing:border-box;}.bani-widgeth2{margin:0020px0;color:#333;font-size:24px;text-align:center;}.bani-widgeth3{margin:0016px0;font-size:16px;}.bani-widget-loading{text-align:center;padding:40px20px;color:#666;}.bani-widget-error{background:#ffebee;color:#c62828;padding:12px;border-radius:4px;margin-bottom:16px;}.bani-widget-calendar{margin-bottom:20px;}.bani-widget-calendar-header{display:grid;grid-template-columns:repeat(7,1fr);gap:4px;margin-bottom:8px;}.bani-widget-calendar-grid{display:grid;grid-template-columns:repeat(7,1fr);gap:4px;margin-bottom:20px;}.bani-widget-calendar-day{aspect-ratio:1;display:flex;align-items:center;justify-content:center;font-size:12px;color:#999;font-weight:bold;}.bani-widget-calendar-date{aspect-ratio:1;display:flex;align-items:center;justify-content:center;border:2pxsolid#ddd;border-radius:4px;cursor:pointer;transition:all0.2s;background:white;font-weight:500;font-size:14px;padding:0;}.bani-widget-calendar-date:hover{border-color:#4caf50;background:#f0f0f0;}.bani-widget-calendar-date.selected{background:#4caf50;color:white;border-color:#4caf50;}.bani-widget-calendar-date.disabled{background:#f5f5f5;color:#ccc;cursor:not-allowed;border-color:#ddd;}.bani-widget-slots{margin-bottom:20px;}.bani-widget-slot{display:flex;justify-content:space-between;align-items:center;padding:12px;margin-bottom:8px;border:2pxsolid#ddd;border-radius:4px;cursor:pointer;transition:all0.2s;background:white;font-size:14px;}.bani-widget-slot:hover{border-color:#4caf50;background:#f9f9f9;}.bani-widget-slot.selected{background:#4caf50;color:white;border-color:#4caf50;}.bani-widget-slot.disabled{background:#f5f5f5;color:#ccc;cursor:not-allowed;border-color:#ddd;}.bani-widget-slot-time{font-weight:500;}.bani-widget-slot-price{text-align:right;font-weight:600;}.bani-widget-form-group{margin-bottom:16px;}.bani-widget-label{display:block;margin-bottom:6px;font-weight:500;color:#333;font-size:14px;}.bani-widget-input{width:100%;padding:10px12px;border:2pxsolid#ddd;border-radius:4px;font-size:14px;font-family:inherit;transition:border-color0.2s;}.bani-widget-input:focus{outline:none;border-color:#4caf50;}.bani-widget-input::placeholder{color:#ccc;}.bani-widget-button{width:100%;padding:12px;background:#4caf50;color:white;border:none;border-radius:4px;font-size:16px;font-weight:600;cursor:pointer;transition:background0.2s;font-family:inherit;}.bani-widget-button:hover{background:#45a049;}.bani-widget-button:disabled{background:#ccc;cursor:not-allowed;}.bani-widget-button-secondary{background:#999;margin-bottom:8px;}.bani-widget-button-secondary:hover{background:#777;}.bani-widget-nav{display:flex;gap:8px;margin-bottom:16px;}.bani-widget-navbutton{flex:1;}.bani-widget-bathhouse-info{background:white;padding:16px;border-radius:4px;margin-bottom:20px;border-left:4pxsolid#4caf50;}.bani-widget-bathhouse-name{font-weight:600;font-size:18px;margin-bottom:8px;}.bani-widget-bathhouse-price{color:#4caf50;font-weight:600;font-size:16px;}.bani-widget-confirmation{text-align:center;padding:20px;}.bani-widget-confirmation-icon{font-size:48px;margin-bottom:16px;}.bani-widget-confirmation-message{font-size:18px;margin-bottom:20px;color:#333;font-weight:600;}.bani-widget-confirmation-details{background:white;padding:16px;border-radius:4px;margin-bottom:20px;text-align:left;}.bani-widget-confirmation-detail{display:flex;justify-content:space-between;padding:8px0;border-bottom:1pxsolid#eee;font-size:14px;}.bani-widget-confirmation-detail:last-child{border-bottom:none;}.bani-widget-confirmation-detailstrong{font-weight:600;}@media(max-width:600px){.bani-widget{max-width:100%;padding:12px;}.bani-widgeth2{font-size:20px;}.bani-widget-input,.bani-widget-button{font-size:16px;}}`;
    document.head.appendChild(style);
  }

  // Widget code
  !function(){"use strict";const t={apiBaseUrl:void 0!==window.BANI_WIDGET_API_URL?window.BANI_WIDGET_API_URL:"https://api.bani.ru",defaultLanguage:"ru",defaultPrimaryColor:"#0f766e",defaultFontFamily:'Manrope, "Avenir Next", "Segoe UI", ui-sans-serif, system-ui, sans-serif',translations:{ru:{title:"Бронирование",selectDate:"Выберите дату",selectTime:"Выберите время",bookingForm:"Данные для бронирования",name:"Имя",phone:"Телефон",email:"Email",guests:"Количество гостей",price:"Цена",pricePerHour:"за час",book:"Забронировать",back:"Назад",confirmation:"Спасибо за бронирование!",bookingId:"Номер бронирования:",dateTime:"Дата и время:",totalPrice:"Итого:",close:"Закрыть",loading:"Загрузка...",error:"Ошибка",noSlotsAvailable:"Нет доступных слотов на эту дату",invalidInput:"Пожалуйста, заполните все поля"},en:{title:"Booking",selectDate:"Select date",selectTime:"Select time",bookingForm:"Booking details",name:"Name",phone:"Phone",email:"Email",guests:"Number of guests",price:"Price",pricePerHour:"per hour",book:"Book",back:"Back",confirmation:"Thank you for booking!",bookingId:"Booking ID:",dateTime:"Date and time:",totalPrice:"Total:",close:"Close",loading:"Loading...",error:"Error",noSlotsAvailable:"No available slots for this date",invalidInput:"Please fill in all fields"}}};class e{constructor(e){var i;this.element=e,this.apiKey=e.getAttribute("data-api-key"),this.primaryColor=(i=e.getAttribute("data-color"),(/^#[0-9a-fA-F]{3,8}$/.test(i)?i:null)||t.defaultPrimaryColor),this.fontFamily=function(t){return t.replace(/[^a-zA-Z0-9s,-]/g,"")}(e.getAttribute("data-font-family")||t.defaultFontFamily),this.language=e.getAttribute("data-language")||t.defaultLanguage,this.t=t.translations[this.language]||t.translations[t.defaultLanguage],this.apiKey?(this.state={view:"date-selection",selectedDate:null,selectedSlot:null,bathhouse:null,slots:[],loading:!1,error:null},this.init()):console.error("Bani Widget: data-api-key attribute is required")}init(){this.setupStyles(),this.render(),this.fetchBathhouse()}setupStyles(){if(!document.getElementById("bani-widget-styles")){const t=document.createElement("style");t.id="bani-widget-styles",t.textContent=this.getStyles(),document.head.appendChild(t)}}getStyles(){return`
        .bani-widget {
          --bani-widget-primary: ${this.primaryColor};
          --bani-widget-primary-strong: ${this.adjustColor(this.primaryColor,-18)};
          --bani-widget-primary-soft: rgba(15, 118, 110, 0.08);
          --bani-widget-accent: #d97706;
          --bani-widget-error: #b42318;
          --bani-widget-text: #16212b;
          --bani-widget-text-soft: #5f6877;
          --bani-widget-border: rgba(15, 23, 42, 0.08);
          --bani-widget-border-strong: rgba(15, 23, 42, 0.12);
          font-family: ${this.fontFamily};
          color: var(--bani-widget-text);
          background:
            radial-gradient(circle at top left, rgba(15, 118, 110, 0.10), transparent 34%),
            linear-gradient(180deg, rgba(255, 255, 255, 0.94), rgba(248, 244, 236, 0.84));
          border: 1px solid var(--bani-widget-border);
          border-radius: 28px;
          padding: 24px;
          max-width: 500px;
          box-shadow: 0 24px 64px rgba(15, 23, 42, 0.10);
          backdrop-filter: blur(18px);
        }
        .bani-widget * { box-sizing: border-box; }
        .bani-widget h2 { margin: 0 0 20px 0; color: var(--bani-widget-text); font-size: 24px; line-height: 1.15; font-weight: 800; letter-spacing: 0; text-align: center; }
        .bani-widget h3 { margin: 0 0 16px 0; font-size: 16px; font-weight: 800; letter-spacing: 0; color: var(--bani-widget-text); }
        .bani-widget-loading { text-align: center; padding: 40px 20px; color: var(--bani-widget-text-soft); font-weight: 600; }
        .bani-widget-error { background: rgba(180, 35, 24, 0.08); color: var(--bani-widget-error); padding: 14px 16px; border: 1px solid rgba(180, 35, 24, 0.16); border-radius: 18px; margin-bottom: 16px; font-weight: 700; }
        .bani-widget-calendar { margin-bottom: 20px; }
        .bani-widget-calendar-grid { display: grid; grid-template-columns: repeat(7, minmax(0, 1fr)); gap: 6px; margin-bottom: 20px; }
        .bani-widget-calendar-header { display: grid; grid-template-columns: repeat(7, minmax(0, 1fr)); gap: 6px; margin-bottom: 8px; }
        .bani-widget-calendar-day { aspect-ratio: 1; display: flex; align-items: center; justify-content: center; font-size: 11px; color: var(--bani-widget-text-soft); font-weight: 800; letter-spacing: 0.08em; }
        .bani-widget-calendar-date { aspect-ratio: 1; display: flex; align-items: center; justify-content: center; border: 1px solid var(--bani-widget-border-strong); border-radius: 16px; cursor: pointer; transition: all 0.2s ease; background: rgba(255, 253, 248, 0.86); color: var(--bani-widget-text); font-weight: 800; font-size: 14px; padding: 0; font-family: inherit; }
        .bani-widget-calendar-date:hover { border-color: var(--bani-widget-primary); background: rgba(255, 255, 255, 0.96); transform: translateY(-1px); box-shadow: 0 12px 24px rgba(15, 23, 42, 0.08); }
        .bani-widget-calendar-date.disabled { background: rgba(244, 239, 231, 0.82); color: rgba(22, 33, 43, 0.42); cursor: not-allowed; border-color: var(--bani-widget-border); box-shadow: none; transform: none; }
        .bani-widget-calendar-date.selected { background: linear-gradient(135deg, var(--bani-widget-primary), var(--bani-widget-primary-strong)); color: white; border-color: var(--bani-widget-primary); box-shadow: 0 14px 28px rgba(15, 118, 110, 0.20); }
        .bani-widget-slots { display: grid; gap: 10px; margin-bottom: 20px; }
        .bani-widget-slot { width: 100%; display: flex; justify-content: space-between; align-items: center; gap: 12px; padding: 14px 16px; border: 1px solid var(--bani-widget-border-strong); border-radius: 18px; cursor: pointer; transition: all 0.2s ease; background: rgba(255, 253, 248, 0.86); color: var(--bani-widget-text); font-family: inherit; font-size: 14px; text-align: left; }
        .bani-widget-slot:hover { border-color: var(--bani-widget-primary); background: rgba(255, 255, 255, 0.96); transform: translateY(-1px); box-shadow: 0 12px 24px rgba(15, 23, 42, 0.08); }
        .bani-widget-slot.disabled { background: rgba(244, 239, 231, 0.82); color: rgba(22, 33, 43, 0.42); cursor: not-allowed; border-color: var(--bani-widget-border); transform: none; box-shadow: none; }
        .bani-widget-slot.selected { background: linear-gradient(135deg, var(--bani-widget-primary), var(--bani-widget-primary-strong)); color: white; border-color: var(--bani-widget-primary); box-shadow: 0 14px 28px rgba(15, 118, 110, 0.20); }
        .bani-widget-slot-time { font-weight: 800; }
        .bani-widget-slot-price { text-align: right; font-weight: 800; }
        .bani-widget-form-group { margin-bottom: 16px; }
        .bani-widget-label { display: block; margin-bottom: 8px; font-weight: 700; color: var(--bani-widget-text); font-size: 13px; }
        .bani-widget-input { width: 100%; min-height: 50px; padding: 12px 14px; border: 1px solid rgba(15, 23, 42, 0.14); border-radius: 16px; font-size: 15px; font-family: inherit; color: var(--bani-widget-text); background: linear-gradient(180deg, rgba(255, 255, 255, 0.98), rgba(251, 247, 240, 0.96)); box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.88), 0 10px 24px rgba(15, 23, 42, 0.04); transition: border-color 0.2s ease, box-shadow 0.2s ease; }
        .bani-widget-input:focus { outline: none; border-color: rgba(15, 118, 110, 0.52); box-shadow: 0 0 0 4px rgba(15, 118, 110, 0.10), 0 18px 32px rgba(15, 23, 42, 0.08); }
        .bani-widget-input::placeholder { color: rgba(95, 104, 119, 0.72); }
        .bani-widget-button { width: 100%; min-height: 44px; padding: 12px 16px; background: linear-gradient(135deg, var(--bani-widget-primary), var(--bani-widget-primary-strong)); color: white; border: none; border-radius: 999px; font-size: 15px; font-weight: 800; cursor: pointer; transition: transform 0.2s ease, box-shadow 0.2s ease; font-family: inherit; box-shadow: 0 14px 28px rgba(15, 118, 110, 0.20); }
        .bani-widget-button:hover { transform: translateY(-1px); box-shadow: 0 18px 30px rgba(15, 118, 110, 0.24); }
        .bani-widget-button:disabled { color: rgba(22, 33, 43, 0.42); background: linear-gradient(180deg, rgba(255, 255, 255, 0.82), rgba(245, 241, 234, 0.96)); cursor: not-allowed; box-shadow: none; transform: none; }
        .bani-widget-button-secondary { background: rgba(255, 255, 255, 0.82); color: var(--bani-widget-text); border: 1px solid var(--bani-widget-border-strong); box-shadow: none; margin-bottom: 8px; }
        .bani-widget-button-secondary:hover { background: rgba(255, 255, 255, 0.96); box-shadow: 0 12px 24px rgba(15, 23, 42, 0.08); }
        .bani-widget-confirmation { text-align: center; padding: 12px 0 0; }
        .bani-widget-confirmation-icon { width: 64px; height: 64px; display: inline-flex; align-items: center; justify-content: center; margin-bottom: 16px; border-radius: 22px; background: rgba(15, 118, 110, 0.10); color: var(--bani-widget-primary); font-size: 38px; font-weight: 900; }
        .bani-widget-confirmation-message { font-size: 18px; margin-bottom: 20px; color: var(--bani-widget-text); font-weight: 800; }
        .bani-widget-confirmation-details { background: rgba(255, 253, 248, 0.82); padding: 18px; border: 1px solid var(--bani-widget-border); border-radius: 22px; margin-bottom: 20px; text-align: left; }
        .bani-widget-confirmation-detail { display: flex; justify-content: space-between; gap: 12px; padding: 10px 0; border-bottom: 1px solid var(--bani-widget-border); font-size: 14px; }
        .bani-widget-confirmation-detail:last-child { border-bottom: none; }
        .bani-widget-bathhouse-info { background: rgba(255, 253, 248, 0.82); padding: 18px; border-radius: 22px; margin-bottom: 20px; border: 1px solid var(--bani-widget-border); box-shadow: 0 16px 40px rgba(15, 23, 42, 0.06); }
        .bani-widget-bathhouse-name { font-weight: 800; font-size: 18px; line-height: 1.2; margin-bottom: 8px; color: var(--bani-widget-text); }
        .bani-widget-bathhouse-address, .bani-widget-muted { color: var(--bani-widget-text-soft); font-size: 14px; line-height: 1.5; }
        .bani-widget-bathhouse-address { margin-bottom: 10px; }
        .bani-widget-bathhouse-price { color: var(--bani-widget-primary); font-weight: 800; font-size: 16px; }
        .bani-widget-selection-summary { background: rgba(255, 253, 248, 0.82); padding: 14px 16px; border-radius: 20px; margin-bottom: 16px; border: 1px solid var(--bani-widget-border); }
        .bani-widget-selection-summary strong { color: var(--bani-widget-text); }
        .bani-widget-empty { text-align: center; padding: 40px 20px; color: var(--bani-widget-text-soft); font-weight: 700; }
        .bani-widget-nav { display: flex; gap: 8px; margin-bottom: 16px; }
        .bani-widget-nav button { flex: 1; }
        @media (max-width: 600px) {
          .bani-widget { max-width: 100%; padding: 16px; border-radius: 24px; }
          .bani-widget h2 { font-size: 22px; }
          .bani-widget-calendar-grid, .bani-widget-calendar-header { gap: 4px; }
          .bani-widget-calendar-date { border-radius: 14px; font-size: 13px; }
          .bani-widget-nav { flex-direction: column; }
        }
      `}adjustColor(t,e){const i=parseInt(t.replace("#",""),16),a=Math.round(2.55*e),n=(i>>16)+a,r=(i>>8JS_CONTENT_HERE255)+a,o=(255JS_CONTENT_HEREi)+a;return"#"+(16777216+65536*(n<255?n<1?0:n:255)+256*(r<255?r<1?0:r:255)+(o<255?o<1?0:o:255)).toString(16).slice(1)}async fetchBathhouse(){try{this.state.loading=!0,this.render();const e=await fetch(`${t.apiBaseUrl}/api/v1/widget/${this.apiKey}/bathhouse`);if(!e.ok)throw new Error("Failed to fetch bathhouse info");const i=await e.json();this.state.bathhouse=i.data,this.state.error=null}catch(t){this.state.error=t.message,console.error("Bani Widget Error:",t)}finally{this.state.loading=!1,this.render()}}async fetchSlots(e){try{this.state.loading=!0,this.render();const i=e.toISOString().split("T")[0],a=await fetch(`${t.apiBaseUrl}/api/v1/widget/${this.apiKey}/slots?date=${i}`);if(!a.ok)throw new Error("Failed to fetch available slots");const n=await a.json();this.state.slots=n.data||[],this.state.selectedDate=e,this.state.view="slot-selection",this.state.error=null}catch(t){this.state.error=t.message,console.error("Bani Widget Error:",t)}finally{this.state.loading=!1,this.render()}}async submitBooking(e){try{this.state.loading=!0,this.render();const i=this.state.selectedSlot,a={name:e.name,phone:e.phone,email:e.email,start_time:i.start_time,end_time:i.end_time,guest_count:parseInt(e.guests),comment:e.comment||""},n=await fetch(`${t.apiBaseUrl}/api/v1/widget/${this.apiKey}/booking`,{method:"POST",headers:{"Content-Type":"application/json"},body:JSON.stringify(a)});if(!n.ok){const t=await n.json();throw new Error(t.error?.message||"Failed to create booking")}const r=await n.json();this.state.selectedSlot={...i,...r.data},this.state.view="confirmation",this.state.error=null,this.state.loading=!1,this.render()}catch(t){this.state.error=t.message,console.error("Bani Widget Error:",t),this.state.loading=!1,this.render()}}formatPrice(t){return(t/100).toLocaleString(this.language,{minimumFractionDigits:0,maximumFractionDigits:2})}render(){const t=this.renderCurrentView();this.element.innerHTML=t,this.attachEventListeners()}renderCurrentView(){if(this.state.error)return`<div class="bani-widget"><div class="bani-widget-error">${this.t.error}: ${this.escapeHtml(this.state.error)}</div></div>`;if(this.state.loadingJS_CONTENT_HEREJS_CONTENT_HERE!this.state.bathhouse)return`<div class="bani-widget"><div class="bani-widget-loading">${this.t.loading}</div></div>`;switch(this.state.view){case"date-selection":return this.renderDateSelection();case"slot-selection":return this.renderSlotSelection();case"booking-form":return this.renderBookingForm();case"confirmation":return this.renderConfirmation();default:return""}}escapeHtml(t){const e={"JS_CONTENT_HERE":"JS_CONTENT_HEREamp;","<":"JS_CONTENT_HERElt;",">":"JS_CONTENT_HEREgt;",'"':"JS_CONTENT_HEREquot;","'":"JS_CONTENT_HERE#039;"};return t.replace(/[JS_CONTENT_HERE<>"']/g,t=>e[t])}renderDateSelection(){if(!this.state.bathhouse)return"";const t=new Date;let e='<div class="bani-widget">';return e+=`<h2>${this.t.title}</h2>`,e+=this.renderBathhouseInfo(),e+='<div class="bani-widget-calendar">',e+=`<h3>${this.t.selectDate}</h3>`,e+=this.renderCalendar(t,30),e+="</div>",e+="</div>",e}renderBathhouseInfo(){if(!this.state.bathhouse)return"";const t=this.state.bathhouse;return`
        <div class="bani-widget-bathhouse-info">
          <div class="bani-widget-bathhouse-name">${this.escapeHtml(t.name)}</div>
          <div class="bani-widget-bathhouse-address">${this.escapeHtml(t.address)}</div>
          <div class="bani-widget-bathhouse-price">
            ${this.formatPrice(t.price_per_hour)} ${this.t.pricePerHour}
          </div>
        </div>
      `}renderCalendar(t,e){const i=["ПН","ВТ","СР","ЧТ","ПТ","СБ","ВС"];let a='<div class="bani-widget-calendar-header">';for(let t=0;t<7;t++){a+=`<div class="bani-widget-calendar-day">${"ru"===this.language?i[t]:["Mon","Tue","Wed","Thu","Fri","Sat","Sun"][t]}</div>`}a+="</div>",a+='<div class="bani-widget-calendar-grid">';for(let i=0;i<e;i++){const e=new Date(t);e.setDate(e.getDate()+i);const n=e.toISOString().split("T")[0];a+=`<button
          class="bani-widget-calendar-date ${this.state.selectedDateJS_CONTENT_HEREJS_CONTENT_HEREthis.state.selectedDate.toISOString().split("T")[0]===n?"selected":""}"
          data-date="${n}"
        >
          ${e.getDate()}
        </button>`}return a+="</div>",a}renderSlotSelection(){if(!this.state.bathhouse)return"";let t='<div class="bani-widget">';if(t+=`<h2>${this.t.title}</h2>`,t+=this.renderBathhouseInfo(),t+=`
        <div class="bani-widget-nav">
          <button class="bani-widget-button bani-widget-button-secondary" data-action="back">
            ${this.t.back}
          </button>
        </div>
      `,t+='<div class="bani-widget-selection-summary">',t+=`<strong>${this.t.selectTime}</strong>`,this.state.selectedDate){const e=this.state.selectedDate.toLocaleDateString(this.language);t+=`<div class="bani-widget-muted">${e}</div>`}return t+="</div>",0===this.state.slots.length?(t+='<div class="bani-widget-empty">',t+=this.t.noSlotsAvailable,t+="</div>"):(t+='<div class="bani-widget-slots">',this.state.slots.forEach(e=>{const i=this.state.selectedSlotJS_CONTENT_HEREJS_CONTENT_HEREthis.state.selectedSlot.start_time===e.start_time,a=e.available,n=new Date(e.start_time),r=new Date(e.end_time),o=`${n.getHours().toString().padStart(2,"0")}:${n.getMinutes().toString().padStart(2,"0")} - ${r.getHours().toString().padStart(2,"0")}:${r.getMinutes().toString().padStart(2,"0")}`;var s;t+=`
            <button
              class="bani-widget-slot ${i?"selected":""} ${a?"":"disabled"}"
              data-slot='${s=JSON.stringify(e),s.replace(/JS_CONTENT_HERE/g,"JS_CONTENT_HEREamp;").replace(/'/g,"JS_CONTENT_HERE#39;").replace(/"/g,"JS_CONTENT_HEREquot;").replace(/</g,"JS_CONTENT_HERElt;").replace(/>/g,"JS_CONTENT_HEREgt;")}'
              ${a?"":"disabled"}
            >
              <span class="bani-widget-slot-time">${o}</span>
              <span class="bani-widget-slot-price">${this.formatPrice(e.price)}</span>
            </button>
          `}),t+="</div>",this.state.selectedSlotJS_CONTENT_HEREJS_CONTENT_HERE(t+=`
            <button class="bani-widget-button" data-action="to-form">
              ${this.t.bookingForm}
            </button>
          `)),t+="</div>",t}renderBookingForm(){if(!this.state.bathhouse||!this.state.selectedSlot)return"";const t=this.state.selectedSlot,e=new Date(t.start_time),i=new Date(t.end_time),a=`${e.getHours().toString().padStart(2,"0")}:${e.getMinutes().toString().padStart(2,"0")} - ${i.getHours().toString().padStart(2,"0")}:${i.getMinutes().toString().padStart(2,"0")}`,n=e.toLocaleDateString(this.language);let r='<div class="bani-widget">';return r+=`<h2>${this.t.bookingForm}</h2>`,r+=this.renderBathhouseInfo(),r+=`
        <div class="bani-widget-selection-summary">
          <div class="bani-widget-muted">${this.t.selectTime}</div>
          <div style="font-size: 16px; font-weight: 800; color: var(--bani-widget-text);">${n}</div>
          <div style="font-size: 16px; font-weight: 800; color: var(--bani-widget-text);">${a}</div>
          <div class="bani-widget-muted" style="margin-top: 8px;">
            ${this.t.totalPrice}: <strong>${this.formatPrice(t.price)}</strong>
          </div>
        </div>
      `,r+=`
        <form id="bani-booking-form" style="margin-bottom: 16px;">
          <div class="bani-widget-form-group">
            <label class="bani-widget-label">${this.t.name}</label>
            <input type="text" class="bani-widget-input" name="name" required />
          </div>
          <div class="bani-widget-form-group">
            <label class="bani-widget-label">${this.t.phone}</label>
            <input type="tel" class="bani-widget-input" name="phone" required />
          </div>
          <div class="bani-widget-form-group">
            <label class="bani-widget-label">${this.t.email}</label>
            <input type="email" class="bani-widget-input" name="email" required />
          </div>
          <div class="bani-widget-form-group">
            <label class="bani-widget-label">${this.t.guests}</label>
            <input type="number" class="bani-widget-input" name="guests" min="1" max="${this.state.bathhouse.max_guests}" required />
          </div>
          <div class="bani-widget-nav">
            <button type="button" class="bani-widget-button bani-widget-button-secondary" data-action="back">
              ${this.t.back}
            </button>
            <button type="submit" class="bani-widget-button" ${this.state.loading?"disabled":""}>
              ${this.state.loading?this.t.loading:this.t.book}
            </button>
          </div>
        </form>
        </div>
      `,r}renderConfirmation(){const t=this.state.selectedSlot,e=this.state.bathhouse,i=new Date(t.start_time),a=new Date(t.end_time),n=i.toLocaleDateString(this.language)+" "+i.toLocaleTimeString(this.language,{hour:"2-digit",minute:"2-digit"})+" - "+a.toLocaleTimeString(this.language,{hour:"2-digit",minute:"2-digit"});let r='<div class="bani-widget"><div class="bani-widget-confirmation">';return r+='<div class="bani-widget-confirmation-icon">✓</div>',r+=`<div class="bani-widget-confirmation-message">${this.t.confirmation}</div>`,r+=`
        <div class="bani-widget-confirmation-details">
          <div class="bani-widget-confirmation-detail">
            <span>${this.t.bookingId}</span>
            <strong>${this.escapeHtml(String(t.id))}</strong>
          </div>
          <div class="bani-widget-confirmation-detail">
            <span>${this.escapeHtml(e.name)}</span>
          </div>
          <div class="bani-widget-confirmation-detail">
            <span>${this.t.dateTime}</span>
            <span>${n}</span>
          </div>
          <div class="bani-widget-confirmation-detail">
            <span>${this.t.totalPrice}</span>
            <strong>${this.formatPrice(t.price)}</strong>
          </div>
        </div>
      `,r+=`<button class="bani-widget-button" data-action="close">${this.t.close}</button>`,r+="</div></div>",r}attachEventListeners(){const t=this.element.querySelector(".bani-widget");t?.querySelectorAll("[data-date]").forEach(t=>{t.addEventListener("click",e=>{e.preventDefault();const i=t.getAttribute("data-date"),a=new Date(i+"T00:00:00");this.fetchSlots(a)})}),t?.querySelectorAll("[data-slot]").forEach(t=>{t.addEventListener("click",e=>{e.preventDefault();const i=JSON.parse(t.getAttribute("data-slot"));this.state.selectedSlot=i,this.render()})});const e=t?.querySelector("#bani-booking-form");eJS_CONTENT_HEREJS_CONTENT_HEREe.addEventListener("submit",async t=>{t.preventDefault();const i=new FormData(e);if(!i.get("name")||!i.get("phone")||!i.get("email"))return this.state.error=this.t.invalidInput,void this.render();await this.submitBooking({name:i.get("name"),phone:i.get("phone"),email:i.get("email"),guests:i.get("guests"),comment:i.get("comment")||""})}),t?.querySelectorAll("[data-action]").forEach(t=>{t.addEventListener("click",e=>{e.preventDefault();switch(t.getAttribute("data-action")){case"back":"slot-selection"===this.state.view?(this.state.selectedDate=null,this.state.selectedSlot=null,this.state.slots=[],this.state.view="date-selection"):"booking-form"===this.state.viewJS_CONTENT_HEREJS_CONTENT_HERE(this.state.selectedSlot=null,this.state.view="slot-selection"),this.render();break;case"to-form":this.state.view="booking-form",this.render();break;case"close":this.state.view="date-selection",this.state.selectedDate=null,this.state.selectedSlot=null,this.state.slots=[],this.render()}})})}}function i(){document.querySelectorAll('[id="bani-widget"]').forEach(t=>{t.dataset.initialized||(new e(t),t.dataset.initialized="true")})}"loading"===document.readyState?document.addEventListener("DOMContentLoaded",i):i(),window.BaniWidget=e}();
})();
