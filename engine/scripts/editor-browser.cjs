"use strict";

const fs = require("node:fs");
const path = require("node:path");
const { chromium } = require("playwright");
const { ARTIFACT_ID_PATTERN, CATEGORIES, SEVERITIES, hostBrowserEnvironment, parseReviewFile, validateFlags, validateReview } = require("./review-common.cjs");

const PROTOCOL = "dawg.editor.v1";
let sequence = 0;

function emit(event) {
    process.stdout.write(`${JSON.stringify({ protocol: PROTOCOL, sequence: ++sequence, ...event })}\n`);
}

function args(values) {
    const result = {};
    for (let index = 0; index < values.length; index += 1) {
        if (!values[index].startsWith("--")) continue;
        const name = values[index].slice(2);
        result[name] = values[index + 1] && !values[index + 1].startsWith("--") ? values[++index] : true;
    }
    return result;
}

function integer(value, name) {
    if (!/^\d+$/.test(String(value))) throw new Error(`${name} must be a non-negative integer`);
    const parsed = Number(value);
    if (!Number.isSafeInteger(parsed)) throw new Error(`${name} is outside the supported range`);
    return parsed;
}

function eventsFromFile(filePath) {
    const events = fs.readFileSync(filePath, "utf8").split("\n").filter(line => line.trim()).map((line, index) => {
        try {
            const event = JSON.parse(line);
            if (!event || typeof event !== "object" || Array.isArray(event)) throw new Error("must be an object");
            return event;
        } catch (error) {
            throw new Error(`invalid rrweb JSONL line ${index + 1}: ${error.message}`);
        }
    });
    if (!events.length) throw new Error("no rrweb events found in input");
    const duration = Math.round(Number(events.at(-1).timestamp) - Number(events[0].timestamp));
    if (!Number.isSafeInteger(duration) || duration < 0) throw new Error("rrweb timestamps cannot determine a valid duration");
    return { events, duration };
}

function viewport(events) {
    const meta = events.find(event => event.type === 4 && event.data
        && Number.isFinite(Number(event.data.width)) && Number.isFinite(Number(event.data.height))
        && event.data.width >= 240 && event.data.width <= 3840 && event.data.height >= 160 && event.data.height <= 2160);
    return meta ? { width: Math.round(meta.data.width), height: Math.round(meta.data.height) } : { width: 1280, height: 720 };
}

function pageHtml() {
    return `<!doctype html><meta charset="utf-8"><title>DAWG Editor</title><link rel="stylesheet" href="/rrweb.css"><link rel="stylesheet" href="/editor.css"><main><section id="view"><div id="stage"><div id="root"></div></div></section><section id="controls"><button id="play">Play</button><button id="back">-10s</button><button id="forward">+10s</button><button id="speed">1x</button><div id="track"><input id="time" type="range" min="0" value="0" step="100"><div id="markers"></div></div><output id="clock">0:00 / 0:00</output></section><aside><header><h1>Review flags</h1><output id="status">Saved</output></header><p><button id="point">Add point</button> <button id="range">Add range</button></p><div id="timeline-legend"><span><i class="legend-point"></i>Point</span><span><i class="legend-range"></i>Range</span><small>Select a marker to reveal its drag handles.</small></div><ol id="flags"></ol><p id="empty">No flags yet.</p><form id="form" hidden><h2 id="form-heading">Edit flag</h2><label>Title <input id="title" maxlength="120" required></label><label>Note <textarea id="note" maxlength="2000"></textarea></label><label>Start (ms) <input id="start" type="number" min="0" step="1" required></label><label>End (ms) <input id="end" type="number" min="0" step="1" placeholder="Point flag"></label><label>Category <select id="category"></select></label><label>Severity <select id="severity"></select></label><p id="error" role="alert"></p><div class="form-actions"><button>Apply changes</button><button id="delete" type="button">Delete flag</button></div></form><div id="undo-row" role="status" aria-live="polite" hidden><span id="delete-status"></span><button id="undo-delete" type="button">Undo</button></div><section id="publish"><label>Saved artifact name (optional) <input id="artifact-name" maxlength="120"></label><small id="artifact-name-hint"></small></section><div id="publication-message" role="status" aria-live="polite" hidden></div><footer><button id="draft">Save draft</button><button id="save">Save artifact</button><button id="cancel">Cancel</button></footer></aside></main><script src="/rrweb.js"></script><script src="/editor.js"></script>`;
}

function styles(size) {
    return `:root{color-scheme:dark;font-family:system-ui,sans-serif}*{box-sizing:border-box}*[hidden]{display:none!important}body{margin:0;background:#141414;color:#f3f3f3}main{display:grid;grid-template:1fr auto/minmax(0,1fr) 320px;height:100vh}#view{display:flex;min-width:0;min-height:0;align-items:center;justify-content:center;overflow:hidden;background:#050505}#stage{width:${size.width}px;height:${size.height}px;flex:0 0 auto;overflow:hidden;transform-origin:center;background:#fff}#root,#root>.replayer-wrapper{width:100%;height:100%}#controls{display:grid;grid-template-columns:auto auto auto auto 1fr auto;gap:8px;padding:10px;border-top:1px solid #444}button,input,textarea,select{font:inherit}button,input,textarea,select{border:1px solid #666;border-radius:4px;background:#292929;color:inherit;padding:5px}button{cursor:pointer}button:hover{background:#3b3b3b}button:disabled{opacity:.45;cursor:not-allowed}#track{position:relative;min-height:38px}#time{position:relative;z-index:1;width:100%;height:38px;margin:0}#markers{position:absolute;z-index:2;inset:0 8px;pointer-events:none}.marker{position:absolute;top:50%;transform:translateY(-50%);border:0;pointer-events:none;opacity:.62;transition:opacity .15s,filter .15s}.marker:hover,.marker:focus-within,.marker.selected{z-index:3;opacity:1}.marker-select{position:absolute;z-index:1;top:50%;left:50%;width:100%;min-width:24px;height:24px;padding:0;transform:translate(-50%,-50%);border:0;background:transparent;pointer-events:auto;cursor:pointer}.marker-select:hover{background:transparent}.marker-select:focus-visible{outline:2px solid #fff;outline-offset:2px}.marker.point{top:25%;width:14px!important;height:18px;transform:translate(-50%,-50%);background:transparent}.marker.point::after{content:"";position:absolute;top:2px;left:5px;width:4px;height:14px;border-radius:2px;background:#58a6ff}.marker.point.selected::after{display:none}.marker.range{top:75%;height:10px;border:1px solid #d8b4fe;background:rgba(192,132,252,.62)}.marker.range.selected{background:rgba(192,132,252,.9);box-shadow:0 0 0 2px rgba(192,132,252,.24)}.flag-handle{position:absolute;z-index:2;top:50%;width:18px;height:18px;padding:0;transform:translate(-50%,-50%);border:2px solid #fff;border-radius:50%;background:#58a6ff;pointer-events:auto;touch-action:none;cursor:ew-resize}.flag-handle.start{left:0}.marker.point .flag-handle.start{left:50%}.marker.range .flag-handle{background:#a855f7}.marker.range .flag-handle.start{transform:translate(-100%,-50%)}.flag-handle.end{left:100%;transform:translate(0,-50%);background:#d8b4fe}.flag-handle:focus-visible{outline:2px solid #77a7ff;outline-offset:2px}aside{grid-row:1/3;grid-column:2;display:flex;flex-direction:column;gap:9px;overflow:auto;padding:14px;border-left:1px solid #444}header{display:flex;justify-content:space-between;gap:8px}h1,h2{margin:0;font-size:1rem}#status{color:#b8d7ff;font-size:.85rem}#timeline-legend{display:flex;flex-wrap:wrap;align-items:center;gap:6px 12px;color:#bbb;font-size:.8rem;line-height:1.4}#timeline-legend span{display:inline-flex;align-items:center;gap:5px;color:#ddd}#timeline-legend small{flex-basis:100%;color:#aaa}.legend-point{width:4px;height:13px;border-radius:2px;background:#58a6ff}.legend-range{width:20px;height:8px;border:1px solid #d8b4fe;border-radius:3px;background:rgba(192,132,252,.62)}#artifact-name-hint{margin:0;color:#bbb;font-size:.8rem;line-height:1.4}#flags{display:grid;gap:7px;margin:0;padding:0;list-style:none}.flag{width:100%;text-align:left}.flag.point{border-left:3px solid #58a6ff}.flag.range{border-left:3px solid #c084fc}.flag[aria-current=true]{border-color:#77a7ff;background:#273a59}.flag span{display:block;color:#ccc;font-size:.82rem}form{display:grid;gap:7px;border-top:1px solid #444;padding-top:10px}label{display:grid;gap:3px;font-size:.9rem}textarea{min-height:60px}#error{min-height:1.2em;margin:0;color:#ff9c9c;font-size:.85rem}.form-actions,#undo-row{display:flex;align-items:center;gap:7px}.form-actions button{flex:1}#undo-row{justify-content:space-between;padding:8px;border:1px solid #555;border-radius:4px;background:#242424;color:#ddd;font-size:.82rem}#undo-row[hidden]{display:none}#publication-message{padding:9px 10px;border:1px solid;border-radius:4px;font-size:.85rem;line-height:1.4}#publication-message.success{border-color:#3f8f62;background:#183425;color:#b8f3cc}#publication-message.error{border-color:#a84d4d;background:#3a1d1d;color:#ffc2c2}#publish{display:grid;gap:4px;margin-top:auto;padding-top:10px;border-top:1px solid #444}footer{display:flex;gap:7px;padding-top:10px;border-top:1px solid #444}#save{background:#255e9d}@media(max-width:800px){main{grid-template:1fr auto auto/1fr}aside{grid-row:3;grid-column:1;max-height:45vh;border-left:0;border-top:1px solid #444}}`;
}

function clientScript(size, categories, severities, publishOptions = { defaultTitle: "Reviewed artifact" }) {
    return `(async()=>{try{const [er,rr]=await Promise.all([fetch('/events.json'),fetch('/review.json')]);if(!er.ok||!rr.ok)throw Error('editor resources could not be loaded');const events=await er.json(),initial=await rr.json(),publish=${JSON.stringify(publishOptions)};let flags=Array.isArray(initial.flags)?initial.flags.slice():[],selectedId=flags[0]?.id||null,playing=false,dirty=false,speedIndex=1,scrubbing=false,lastDeleted=null;const by=id=>document.getElementById(id),root=by('root'),view=by('view'),stage=by('stage'),play=by('play'),back=by('back'),forward=by('forward'),speed=by('speed'),timeline=by('time'),clock=by('clock'),markers=by('markers'),list=by('flags'),empty=by('empty'),form=by('form'),formHeading=by('form-heading'),title=by('title'),note=by('note'),start=by('start'),end=by('end'),category=by('category'),severity=by('severity'),error=by('error'),status=by('status'),artifactName=by('artifact-name'),artifactNameHint=by('artifact-name-hint'),deleteButton=by('delete'),undoRow=by('undo-row'),undoDelete=by('undo-delete'),deleteStatus=by('delete-status'),draftButton=by('draft'),saveButton=by('save'),publicationMessage=by('publication-message');for(const value of ${JSON.stringify(categories)}){const option=document.createElement('option');option.value=value;option.textContent=value;category.append(option)}for(const value of ${JSON.stringify(severities)}){const option=document.createElement('option');option.value=value;option.textContent=value;severity.append(option)}artifactName.placeholder=publish.defaultTitle;artifactNameHint.textContent='Leave blank to use “'+publish.defaultTitle+'”. A new immutable revision is always created.';const replayer=new rrweb.Replayer(events,{root,unpackFn:rrweb.unpack}),duration=Math.max(0,Number.isSafeInteger(publish.durationMs)?publish.durationMs:Math.round(replayer.getMetaData().totalTime)),speeds=[.5,1,1.5,2,4],viewSize=${JSON.stringify(size)};const clamp=value=>Math.max(0,Math.min(duration,Math.round(Number(value)||0))),format=value=>{const seconds=Math.floor(Math.max(0,value)/1000),m=Math.floor(seconds/60)%60,s=seconds%60,h=Math.floor(seconds/3600),base=m+':'+String(s).padStart(2,'0');return h?h+':'+base.padStart(5,'0'):base},selected=()=>flags.find(flag=>flag.id===selectedId)||null,state=()=>({currentTimeMs:clamp(replayer.getCurrentTime()),durationMs:duration,playing,speed:speeds[speedIndex],selectedFlagId:selectedId,flagCount:flags.length,dirty});const update=time=>{if(!scrubbing)timeline.value=String(clamp(time));clock.textContent=format(scrubbing?Number(timeline.value):time)+' / '+format(duration);play.textContent=playing?'Pause':'Play';speed.textContent=speeds[speedIndex]+'x'},resize=()=>{stage.style.transform='scale('+Math.max(.01,Math.min(view.clientWidth/viewSize.width,view.clientHeight/viewSize.height))+')'},seek=value=>{const target=clamp(value),resume=playing;replayer.pause(target);if(resume)replayer.play(target);update(target);return target};const send=async type=>{dirty=true;status.textContent='Unsaved changes';return window.__dawgEditorIntent({type,flags})};const positionMarker=(marker,flag)=>{marker.style.left=Math.max(0,Math.min(100,flag.startOffsetMs/Math.max(1,duration)*100))+'%';marker.style.width=((flag.endOffsetMs===undefined?0:flag.endOffsetMs-flag.startOffsetMs)/Math.max(1,duration)*100)+'%'};const offsetFromPointer=event=>{const bounds=markers.getBoundingClientRect();return clamp((event.clientX-bounds.left)/Math.max(1,bounds.width)*duration)};const setFlagOffset=(flag,edge,requested)=>{if(edge==='end')flag.endOffsetMs=Math.max(flag.startOffsetMs+1,Math.min(duration,requested));else flag.startOffsetMs=Math.min(flag.endOffsetMs===undefined?duration:flag.endOffsetMs-1,requested)};const focusHandle=(id,edge)=>requestAnimationFrame(()=>markers.querySelector('[data-flag-id="'+id+'"][data-edge="'+edge+'"]')?.focus());const beginDrag=(event,flag,edge,marker,timing)=>{event.preventDefault();event.stopPropagation();const handle=event.currentTarget;let moved=false;selectedId=flag.id;handle.setPointerCapture(event.pointerId);const move=pointerEvent=>{moved=true;setFlagOffset(flag,edge,offsetFromPointer(pointerEvent));positionMarker(marker,flag);timing.textContent=flag.endOffsetMs===undefined?format(flag.startOffsetMs):format(flag.startOffsetMs)+' – '+format(flag.endOffsetMs);start.value=String(flag.startOffsetMs);end.value=flag.endOffsetMs===undefined?'':String(flag.endOffsetMs);seek(edge==='end'?flag.endOffsetMs:flag.startOffsetMs);dirty=true;status.textContent='Unsaved changes';error.textContent=''};const finish=()=>{handle.removeEventListener('pointermove',move);handle.removeEventListener('pointerup',finish);handle.removeEventListener('pointercancel',finish);if(moved){render();focusHandle(flag.id,edge);void send('draftChanged')}};handle.addEventListener('pointermove',move);handle.addEventListener('pointerup',finish);handle.addEventListener('pointercancel',finish)};const nudgeFlag=(event,flag,edge)=>{if(event.key!=='ArrowLeft'&&event.key!=='ArrowRight')return;event.preventDefault();event.stopPropagation();const direction=event.key==='ArrowLeft'?-1:1,step=event.shiftKey?1000:100,current=edge==='end'?flag.endOffsetMs:flag.startOffsetMs;setFlagOffset(flag,edge,current+direction*step);seek(edge==='end'?flag.endOffsetMs:flag.startOffsetMs);render();focusHandle(flag.id,edge);void send('draftChanged')};const selectFlag=flag=>{selectedId=flag.id;error.textContent='';seek(flag.startOffsetMs);render()};const render=()=>{flags.sort((left,right)=>left.startOffsetMs-right.startOffsetMs||left.id.localeCompare(right.id));list.replaceChildren();markers.replaceChildren();empty.hidden=flags.length>0;const current=selected();form.hidden=!current;deleteButton.disabled=!current;draftButton.disabled=flags.length===0;saveButton.disabled=flags.length===0;for(const flag of flags){const isRange=flag.endOffsetMs!==undefined,type=isRange?'range':'point',isSelected=flag.id===selectedId,item=document.createElement('li'),button=document.createElement('button'),label=document.createElement('strong'),timing=document.createElement('span');button.type='button';button.className='flag '+type;button.setAttribute('aria-current',String(isSelected));label.textContent=flag.title;timing.textContent=(isRange?'Range · ':'Point · ')+(isRange?format(flag.startOffsetMs)+' – '+format(flag.endOffsetMs):format(flag.startOffsetMs));button.append(label,timing);button.onclick=event=>{selectFlag(flag);if(event.detail===0)focusHandle(flag.id,'start')};item.append(button);list.append(item);const marker=document.createElement('div'),markerSelect=document.createElement('button');marker.className='marker '+type+(isSelected?' selected':'');markerSelect.type='button';markerSelect.className='marker-select';markerSelect.setAttribute('aria-label','Select '+type+' flag: '+flag.title);markerSelect.title=(isRange?'Range':'Point')+': '+flag.title;markerSelect.onclick=event=>{selectFlag(flag);if(event.detail===0)focusHandle(flag.id,'start')};marker.append(markerSelect);positionMarker(marker,flag);if(isSelected){const startHandle=document.createElement('button');startHandle.type='button';startHandle.className='flag-handle start';startHandle.dataset.flagId=flag.id;startHandle.dataset.edge='start';startHandle.setAttribute('role','slider');startHandle.setAttribute('aria-label','Adjust start of '+flag.title);startHandle.setAttribute('aria-valuemin','0');startHandle.setAttribute('aria-valuemax',String(isRange?flag.endOffsetMs-1:duration));startHandle.setAttribute('aria-valuenow',String(flag.startOffsetMs));startHandle.title='Drag, or use arrow keys. Hold Shift for 1 second steps.';startHandle.onclick=event=>event.stopPropagation();startHandle.onkeydown=event=>nudgeFlag(event,flag,'start');startHandle.addEventListener('pointerdown',event=>beginDrag(event,flag,'start',marker,timing));marker.append(startHandle);if(isRange){const endHandle=document.createElement('button');endHandle.type='button';endHandle.className='flag-handle end';endHandle.dataset.flagId=flag.id;endHandle.dataset.edge='end';endHandle.setAttribute('role','slider');endHandle.setAttribute('aria-label','Adjust end of '+flag.title);endHandle.setAttribute('aria-valuemin',String(flag.startOffsetMs+1));endHandle.setAttribute('aria-valuemax',String(duration));endHandle.setAttribute('aria-valuenow',String(flag.endOffsetMs));endHandle.title='Drag, or use arrow keys. Hold Shift for 1 second steps.';endHandle.onclick=event=>event.stopPropagation();endHandle.onkeydown=event=>nudgeFlag(event,flag,'end');endHandle.addEventListener('pointerdown',event=>beginDrag(event,flag,'end',marker,timing));marker.append(endHandle)}}markers.append(marker)}if(current){const currentIndex=flags.findIndex(flag=>flag.id===current.id),currentType=current.endOffsetMs===undefined?'point':'range';formHeading.textContent='Edit '+currentType+' · '+(currentIndex+1)+' of '+flags.length;title.value=current.title;note.value=current.note||'';start.value=String(current.startOffsetMs);end.value=current.endOffsetMs===undefined?'':String(current.endOffsetMs);end.min=String(current.startOffsetMs+1);category.value=current.category;severity.value=current.severity}};const makeId=()=>{const bytes=new Uint8Array(16);crypto.getRandomValues(bytes);return Array.from(bytes,byte=>byte.toString(16).padStart(2,'0')).join('')};const add=async isRange=>{const now=clamp(replayer.getCurrentTime()),flag={id:makeId(),startOffsetMs:now,title:'New review flag',note:'',category:'note',severity:'info'};if(isRange)flag.endOffsetMs=Math.min(duration,now+1000);if(isRange&&flag.endOffsetMs<=now){error.textContent='A range needs time after its start';return}flags.push(flag);selectedId=flag.id;render();await send('draftChanged')};const apply=async()=>{const old=selected();if(!old)return;const newStart=integer(start.value),newEnd=end.value===''?undefined:integer(end.value);if(!title.value.trim()||title.value!==title.value.trim()){error.textContent='Title is required and cannot have leading or trailing spaces';return}if(title.value.length>120){error.textContent='Title must be at most 120 characters';return}if(note.value.length>2000){error.textContent='Note must be at most 2000 characters';return}if(!Number.isSafeInteger(newStart)||newStart<0||newStart>duration){error.textContent='Start must be a whole number between 0 and '+duration+' ms';return}if(newEnd!==undefined&&(!Number.isSafeInteger(newEnd)||newEnd<=newStart||newEnd>duration)){error.textContent='End must be a whole number after start and no later than '+duration+' ms';return}Object.assign(old,{title:title.value,note:note.value,startOffsetMs:newStart,category:category.value,severity:severity.value});if(newEnd===undefined)delete old.endOffsetMs;else old.endOffsetMs=newEnd;error.textContent='';render();await send('draftChanged')};const integer=value=>/^\\d+$/.test(value)&&Number.isSafeInteger(Number(value))?Number(value):NaN;timeline.max=String(duration);start.max=String(duration);end.max=String(duration);replayer.on(rrweb.ReplayerEvents.Start,()=>{playing=true;update(replayer.getCurrentTime())});replayer.on(rrweb.ReplayerEvents.Pause,()=>{playing=false;update(replayer.getCurrentTime())});replayer.on(rrweb.ReplayerEvents.Resize,dimension=>{if(Number.isFinite(Number(dimension?.width))&&Number.isFinite(Number(dimension?.height))){viewSize.width=Math.round(dimension.width);viewSize.height=Math.round(dimension.height);resize()}});play.onclick=()=>playing?replayer.pause():replayer.play(clamp(replayer.getCurrentTime()));back.onclick=()=>seek(replayer.getCurrentTime()-10000);forward.onclick=()=>seek(replayer.getCurrentTime()+10000);speed.onclick=()=>{speedIndex=(speedIndex+1)%speeds.length;replayer.setConfig({speed:speeds[speedIndex]});update(replayer.getCurrentTime())};timeline.onpointerdown=()=>{scrubbing=true};timeline.oninput=()=>update(Number(timeline.value));timeline.onchange=()=>{scrubbing=false;seek(timeline.value)};by('point').onclick=()=>void add(false);by('range').onclick=()=>void add(true);form.onsubmit=event=>{event.preventDefault();void apply()};deleteButton.onclick=()=>{const flag=selected(),index=flags.findIndex(item=>item.id===selectedId);if(!flag||index<0)return;lastDeleted={flag,index};flags.splice(index,1);selectedId=(flags[Math.min(index,flags.length-1)]||null)?.id||null;deleteStatus.textContent='Deleted “'+flag.title+'”.';undoRow.hidden=false;error.textContent='';render();requestAnimationFrame(()=>undoDelete.focus());void send('draftChanged')};undoDelete.onclick=()=>{if(!lastDeleted)return;flags.splice(Math.min(lastDeleted.index,flags.length),0,lastDeleted.flag);selectedId=lastDeleted.flag.id;lastDeleted=null;undoRow.hidden=true;render();seek(selected().startOffsetMs);focusHandle(selectedId,'start');void send('draftChanged')};draftButton.onclick=()=>void window.__dawgEditorIntent({type:'saveDraft',flags});saveButton.onclick=()=>{const artifactTitle=artifactName.value.trim();if(Array.from(artifactTitle).length>120){error.textContent='Saved artifact name must be at most 120 characters';return}publicationMessage.hidden=true;publicationMessage.className='';status.textContent='Publishing…';void window.__dawgEditorIntent({type:'saveArtifact',flags,artifactTitle}).then(()=>{status.textContent='Waiting for publication…'}).catch(value=>{status.textContent='Save failed';error.textContent=String(value.message||value)})};by('cancel').onclick=()=>void window.__dawgEditorIntent({type:'close'});new ResizeObserver(resize).observe(view);setInterval(()=>update(replayer.getCurrentTime()),100);replayer.pause(0);resize();render();update(0);const showPublicationResult=result=>{publicationMessage.hidden=false;publicationMessage.className=result.success?'success':'error';if(result.success){const savedName=result.artifactTitle?' “'+result.artifactTitle+'”':'';publicationMessage.textContent='Artifact saved as a new artifact'+savedName+'. It is safe to close the Editor browser.';status.textContent='Published';dirty=false}else{publicationMessage.textContent='Artifact could not be saved: '+(result.message||'Unknown publication error');status.textContent='Save failed'}};window.__DAWG_EDITOR__={state,publicationConfirmed:result=>showPublicationResult({...result,success:true}),publicationFailed:result=>showPublicationResult({...result,success:false}),play:()=>replayer.play(clamp(replayer.getCurrentTime())),pause:()=>replayer.pause(),seek,setSpeed:value=>{const index=speeds.indexOf(Number(value));if(index<0)throw Error('unsupported speed');speedIndex=index;replayer.setConfig({speed:value});update(replayer.getCurrentTime())},flags:()=>flags,replaceFlags:value=>{const previousId=selectedId;flags=value;selectedId=flags.some(flag=>flag.id===previousId)?previousId:flags[0]?.id||null;lastDeleted=null;undoRow.hidden=true;deleteStatus.textContent='';render()}}}catch(error){window.__DAWG_EDITOR_ERROR__=error.stack||error.message;console.error(error)}})();`;
}

function installControlChannel(handle) {
    let buffered = "";
    let closed = false;
    const onData = chunk => {
        if (closed) return;
        buffered += chunk;
        if (Buffer.byteLength(buffered, "utf8") > 1024 * 1024) {
            buffered = "";
            emit({ type: "validationError", message: "editor command exceeds 1048576 bytes" });
            return;
        }
        const lines = buffered.split("\n");
        buffered = lines.pop();
        for (const line of lines) {
            if (!line.trim()) continue;
            let command;
            try {
                command = JSON.parse(line);
            } catch (error) {
                emit({ type: "validationError", message: `invalid JSON command: ${error.message}` });
                continue;
            }
            handle(command).catch(error => {
                if (!closed) emit({ type: "validationError", commandId: command?.id || null, message: error.message });
            });
        }
    };
    const onEnd = () => { closed = true; };
    process.stdin.setEncoding("utf8");
    process.stdin.on("data", onData);
    process.stdin.on("end", onEnd);
    process.stdin.resume();
    return () => {
        closed = true;
        process.stdin.off("data", onData);
        process.stdin.off("end", onEnd);
        process.stdin.pause();
    };
}

function metadata(options, initial) {
    if (initial) return { ...initial, flags: undefined };
    if (!ARTIFACT_ID_PATTERN.test(String(options["root-artifact-id"]))) throw new Error("--root-artifact-id must be a DAWG SHA-256 artifact ID");
    if (!ARTIFACT_ID_PATTERN.test(String(options["parent-artifact-id"]))) throw new Error("--parent-artifact-id must be a DAWG SHA-256 artifact ID");
    return { formatVersion: "1", kind: "flagged", rootArtifactId: options["root-artifact-id"], parentArtifactId: options["parent-artifact-id"], revision: integer(options.revision, "--revision") };
}

async function main() {
    const options = args(process.argv.slice(2));
    if (typeof options["rrweb-input"] !== "string" || options["rrweb-input"].length === 0) throw new Error("missing --rrweb-input path");
    const { events, duration } = eventsFromFile(options["rrweb-input"]);
    const initial = parseReviewFile(options["review-file"], duration, { allowEmpty: true });
    const reviewMetadata = metadata(options, initial);
    const initialFlags = initial?.flags || [];
    const rrwebRoot = path.dirname(require.resolve("rrweb"));
    const rrwebBundle = fs.readFileSync(path.join(rrwebRoot, "rrweb.umd.cjs"), "utf8");
    const rrwebCss = fs.readFileSync(path.join(rrwebRoot, "style.css"), "utf8");
    const recordedViewport = viewport(events);
    const launchOptions = { headless: false, env: hostBrowserEnvironment(), args: [`--window-size=${recordedViewport.width},${Math.min(2400, recordedViewport.height + 180)}`] };
    if (process.env.DAWG_CHROMIUM_EXECUTABLE_PATH) launchOptions.executablePath = process.env.DAWG_CHROMIUM_EXECUTABLE_PATH;
    const browser = await chromium.launch(launchOptions);
    try {
        const context = await browser.newContext({ viewport: null, ignoreHTTPSErrors: true, ...(options["color-scheme"] === "dark" || options["color-scheme"] === "light" ? { colorScheme: options["color-scheme"] } : {}) });
        const page = await context.newPage();
        let closing = false;
        const currentReview = flags => validateReview({ ...reviewMetadata, reviewedAt: new Date().toISOString(), flags }, duration);
        const acceptFlags = flags => {
            if (!Array.isArray(flags)) throw new Error("flags must be an array");
            return flags.length === 0 ? [] : validateFlags(flags, duration);
        };
        const state = async () => page.evaluate(() => window.__DAWG_EDITOR__?.state());
        const handle = async (message, source = "command") => {
            if (!message || typeof message !== "object" || Array.isArray(message) || typeof message.type !== "string") throw new Error("message type is required");
            if (message.protocol && message.protocol !== PROTOCOL) throw new Error("unsupported editor protocol");
            if (!["getState", "play", "pause", "seek", "setSpeed", "addFlag", "updateFlag", "deleteFlag", "saveDraft", "discardDraft", "saveArtifact", "close", "draftChanged", "publicationConfirmed", "publicationFailed"].includes(message.type)) throw new Error("unsupported editor command");
            if (message.type === "seek" && (!Number.isSafeInteger(message.offsetMs) || message.offsetMs < 0)) throw new Error("seek requires a non-negative integer offsetMs");
            if (message.type === "setSpeed" && ![0.5, 1, 1.5, 2, 4].includes(message.speed)) throw new Error("unsupported editor speed");
            if (message.type === "publicationConfirmed" && message.artifactTitle !== undefined && typeof message.artifactTitle !== "string") throw new Error("publicationConfirmed artifactTitle must be a string");
            if (message.type === "publicationFailed" && typeof message.message !== "string") throw new Error("publicationFailed requires a message");
            if (["draftChanged", "saveDraft", "saveArtifact"].includes(message.type)) acceptFlags(message.flags);
            if (message.type === "saveArtifact") {
                if (message.artifactTitle !== undefined && typeof message.artifactTitle !== "string") throw new Error("artifactTitle must be a string");
                const artifactTitle = String(message.artifactTitle || "").trim();
                if (Array.from(artifactTitle).length > 120) throw new Error("artifactTitle must be at most 120 characters");
                const review = currentReview(message.flags);
                emit({ type: "artifactSaved", commandId: message.id || null, review, artifactTitle, source });
                return { review, artifactTitle };
            }
            if (message.type === "saveDraft") { emit({ type: "draftSaved", commandId: message.id || null, flags: acceptFlags(message.flags), source }); return {}; }
            if (message.type === "draftChanged") { emit({ type: "draftChanged", flags: acceptFlags(message.flags), source }); return {}; }
            if (message.type === "discardDraft") { await page.evaluate(() => window.__DAWG_EDITOR__.replaceFlags([])); emit({ type: "draftChanged", flags: [], source }); return {}; }
            if (message.type === "close") { closing = true; emit({ type: "closed", commandId: message.id || null, source }); setTimeout(() => page.close().catch(() => {}), 0); return {}; }
            if (["addFlag", "updateFlag", "deleteFlag"].includes(message.type)) {
                const nextFlags = await page.evaluate(() => window.__DAWG_EDITOR__?.flags());
                if (!Array.isArray(nextFlags)) throw new Error("editor is not ready");
                if (message.type === "addFlag") {
                    if (!message.flag || typeof message.flag !== "object" || Array.isArray(message.flag)) throw new Error("addFlag requires a flag object");
                    nextFlags.push(message.flag);
                } else if (message.type === "updateFlag") {
                    if (!message.flag || typeof message.flag !== "object" || Array.isArray(message.flag) || typeof message.flag.id !== "string") throw new Error("updateFlag requires a complete flag object");
                    const index = nextFlags.findIndex(flag => flag.id === message.flag.id);
                    if (index < 0) throw new Error("cannot update an unknown flag");
                    nextFlags[index] = message.flag;
                } else {
                    if (typeof message.flagId !== "string") throw new Error("deleteFlag requires flagId");
                    const index = nextFlags.findIndex(flag => flag.id === message.flagId);
                    if (index < 0) throw new Error("cannot delete an unknown flag");
                    nextFlags.splice(index, 1);
                }
                const normalized = acceptFlags(nextFlags);
                await page.evaluate(flags => window.__DAWG_EDITOR__.replaceFlags(flags), normalized);
                emit({ type: "draftChanged", commandId: message.id || null, flags: normalized, source });
                return state();
            }
            const result = await page.evaluate(command => { const editor = window.__DAWG_EDITOR__; if (!editor) throw new Error("editor is not ready"); if (command.type === "play") editor.play(); else if (command.type === "pause") editor.pause(); else if (command.type === "seek") editor.seek(command.offsetMs); else if (command.type === "setSpeed") editor.setSpeed(command.speed); else if (command.type === "publicationConfirmed") editor.publicationConfirmed(command); else if (command.type === "publicationFailed") editor.publicationFailed(command); return editor.state(); }, message);
            if (source === "command") emit({ type: "state", commandId: message.id || null, ...result });
            return result;
        };
        await page.exposeFunction("__dawgEditorIntent", message => handle(message, "ui").catch(error => { emit({ type: "validationError", message: error.message }); throw error; }));
        const resources = new Map([["/", ["text/html", pageHtml()]],["/editor.css", ["text/css", styles(recordedViewport)]],["/editor.js", ["text/javascript", clientScript(recordedViewport, CATEGORIES, SEVERITIES, { defaultTitle: options["default-artifact-title"] || `Reviewed artifact ${reviewMetadata.revision}`, durationMs: duration })]],["/rrweb.js", ["text/javascript", rrwebBundle]],["/rrweb.css", ["text/css", rrwebCss]],["/events.json", ["application/json", JSON.stringify(events)]],["/review.json", ["application/json", JSON.stringify({ flags: initialFlags })]]]);
        await page.route("**/*", route => { const url = new URL(route.request().url()); if (url.origin !== "http://dawg-editor.local") return route.abort(); const resource = resources.get(url.pathname); return resource ? route.fulfill({ contentType: resource[0], body: resource[1] }) : route.fulfill({ status: 404, contentType: "text/plain", body: "Not found" }); });
        await page.goto("http://dawg-editor.local/", { waitUntil: "load" });
        await page.waitForFunction(() => window.__DAWG_EDITOR__ || window.__DAWG_EDITOR_ERROR__);
        const initialState = await state();
        if (!initialState) throw new Error(await page.evaluate(() => window.__DAWG_EDITOR_ERROR__ || "editor failed to initialize"));
        emit({ type: "ready", ...initialState });
        const closeControlChannel = installControlChannel(handle);
        try {
            await Promise.race([new Promise(resolve => browser.once("disconnected", resolve)), new Promise(resolve => page.once("close", resolve))]);
        } finally {
            closeControlChannel();
        }
        if (!closing) emit({ type: "closed" });
    } finally { await browser.close().catch(() => {}); }
}

if (require.main === module) {
    main().catch(error => { process.stderr.write(`${error.stack || error.message}\n`); process.exitCode = 1; });
}

module.exports = { clientScript, eventsFromFile, pageHtml, parseArguments: args, styles };
