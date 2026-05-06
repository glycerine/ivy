/**
* @vue/shared v3.5.33
* (c) 2018-present Yuxi (Evan) You and Vue contributors
* @license MIT
**/function Ns(e){const t=Object.create(null);for(const n of e.split(","))t[n]=1;return n=>n in t}const se={},Tt=[],Ue=()=>{},Ri=()=>!1,Hn=e=>e.charCodeAt(0)===111&&e.charCodeAt(1)===110&&(e.charCodeAt(2)>122||e.charCodeAt(2)<97),On=e=>e.startsWith("onUpdate:"),ye=Object.assign,Bs=(e,t)=>{const n=e.indexOf(t);n>-1&&e.splice(n,1)},lr=Object.prototype.hasOwnProperty,ee=(e,t)=>lr.call(e,t),O=Array.isArray,Lt=e=>gn(e)==="[object Map]",Mi=e=>gn(e)==="[object Set]",si=e=>gn(e)==="[object Date]",$=e=>typeof e=="function",ce=e=>typeof e=="string",Le=e=>typeof e=="symbol",te=e=>e!==null&&typeof e=="object",Vi=e=>(te(e)||$(e))&&$(e.then)&&$(e.catch),Gi=Object.prototype.toString,gn=e=>Gi.call(e),cr=e=>gn(e).slice(8,-1),Hi=e=>gn(e)==="[object Object]",jn=e=>ce(e)&&e!=="NaN"&&e[0]!=="-"&&""+parseInt(e,10)===e,qt=Ns(",key,ref,ref_for,ref_key,onVnodeBeforeMount,onVnodeMounted,onVnodeBeforeUpdate,onVnodeUpdated,onVnodeBeforeUnmount,onVnodeUnmounted"),$n=e=>{const t=Object.create(null);return n=>t[n]||(t[n]=e(n))},dr=/-\w/g,we=$n(e=>e.replace(dr,t=>t.slice(1).toUpperCase())),ur=/\B([A-Z])/g,St=$n(e=>e.replace(ur,"-$1").toLowerCase()),zn=$n(e=>e.charAt(0).toUpperCase()+e.slice(1)),rs=$n(e=>e?`on${zn(e)}`:""),We=(e,t)=>!Object.is(e,t),os=(e,...t)=>{for(let n=0;n<e.length;n++)e[n](...t)},Oi=(e,t,n,i=!1)=>{Object.defineProperty(e,t,{configurable:!0,enumerable:!1,writable:i,value:n})},hr=e=>{const t=parseFloat(e);return isNaN(t)?e:t};let ii;const Wn=()=>ii||(ii=typeof globalThis<"u"?globalThis:typeof self<"u"?self:typeof window<"u"?window:typeof global<"u"?global:{});function Te(e){if(O(e)){const t={};for(let n=0;n<e.length;n++){const i=e[n],s=ce(i)?vr(i):Te(i);if(s)for(const a in s)t[a]=s[a]}return t}else if(ce(e)||te(e))return e}const fr=/;(?![^(]*\))/g,pr=/:([^]+)/,gr=/\/\*[^]*?\*\//g;function vr(e){const t={};return e.replace(gr,"").split(fr).forEach(n=>{if(n){const i=n.split(pr);i.length>1&&(t[i[0].trim()]=i[1].trim())}}),t}function R(e){let t="";if(ce(e))t=e;else if(O(e))for(let n=0;n<e.length;n++){const i=R(e[n]);i&&(t+=i+" ")}else if(te(e))for(const n in e)e[n]&&(t+=n+" ");return t.trim()}const mr="itemscope,allowfullscreen,formnovalidate,ismap,nomodule,novalidate,readonly",yr=Ns(mr);function ji(e){return!!e||e===""}function br(e,t){if(e.length!==t.length)return!1;let n=!0;for(let i=0;n&&i<e.length;i++)n=Fs(e[i],t[i]);return n}function Fs(e,t){if(e===t)return!0;let n=si(e),i=si(t);if(n||i)return n&&i?e.getTime()===t.getTime():!1;if(n=Le(e),i=Le(t),n||i)return e===t;if(n=O(e),i=O(t),n||i)return n&&i?br(e,t):!1;if(n=te(e),i=te(t),n||i){if(!n||!i)return!1;const s=Object.keys(e).length,a=Object.keys(t).length;if(s!==a)return!1;for(const r in e){const o=e.hasOwnProperty(r),l=t.hasOwnProperty(r);if(o&&!l||!o&&l||!Fs(e[r],t[r]))return!1}}return String(e)===String(t)}const $i=e=>!!(e&&e.__v_isRef===!0),ie=e=>ce(e)?e:e==null?"":O(e)||te(e)&&(e.toString===Gi||!$(e.toString))?$i(e)?ie(e.value):JSON.stringify(e,zi,2):String(e),zi=(e,t)=>$i(t)?zi(e,t.value):Lt(t)?{[`Map(${t.size})`]:[...t.entries()].reduce((n,[i,s],a)=>(n[ls(i,a)+" =>"]=s,n),{})}:Mi(t)?{[`Set(${t.size})`]:[...t.values()].map(n=>ls(n))}:Le(t)?ls(t):te(t)&&!O(t)&&!Hi(t)?String(t):t,ls=(e,t="")=>{var n;return Le(e)?`Symbol(${(n=e.description)!=null?n:t})`:e};/**
* @vue/reactivity v3.5.33
* (c) 2018-present Yuxi (Evan) You and Vue contributors
* @license MIT
**/let fe;class Wi{constructor(t=!1){this.detached=t,this._active=!0,this._on=0,this.effects=[],this.cleanups=[],this._isPaused=!1,this.__v_skip=!0,this.parent=fe,!t&&fe&&(this.index=(fe.scopes||(fe.scopes=[])).push(this)-1)}get active(){return this._active}pause(){if(this._active){this._isPaused=!0;let t,n;if(this.scopes)for(t=0,n=this.scopes.length;t<n;t++)this.scopes[t].pause();for(t=0,n=this.effects.length;t<n;t++)this.effects[t].pause()}}resume(){if(this._active&&this._isPaused){this._isPaused=!1;let t,n;if(this.scopes)for(t=0,n=this.scopes.length;t<n;t++)this.scopes[t].resume();for(t=0,n=this.effects.length;t<n;t++)this.effects[t].resume()}}run(t){if(this._active){const n=fe;try{return fe=this,t()}finally{fe=n}}}on(){++this._on===1&&(this.prevScope=fe,fe=this)}off(){if(this._on>0&&--this._on===0){if(fe===this)fe=this.prevScope;else{let t=fe;for(;t;){if(t.prevScope===this){t.prevScope=this.prevScope;break}t=t.prevScope}}this.prevScope=void 0}}stop(t){if(this._active){this._active=!1;let n,i;for(n=0,i=this.effects.length;n<i;n++)this.effects[n].stop();for(this.effects.length=0,n=0,i=this.cleanups.length;n<i;n++)this.cleanups[n]();if(this.cleanups.length=0,this.scopes){for(n=0,i=this.scopes.length;n<i;n++)this.scopes[n].stop(!0);this.scopes.length=0}if(!this.detached&&this.parent&&!t){const s=this.parent.scopes.pop();s&&s!==this&&(this.parent.scopes[this.index]=s,s.index=this.index)}this.parent=void 0}}}function Ui(e){return new Wi(e)}function Ki(){return fe}function _r(e,t=!1){fe&&fe.cleanups.push(e)}let oe;const cs=new WeakSet;class qi{constructor(t){this.fn=t,this.deps=void 0,this.depsTail=void 0,this.flags=5,this.next=void 0,this.cleanup=void 0,this.scheduler=void 0,fe&&fe.active&&fe.effects.push(this)}pause(){this.flags|=64}resume(){this.flags&64&&(this.flags&=-65,cs.has(this)&&(cs.delete(this),this.trigger()))}notify(){this.flags&2&&!(this.flags&32)||this.flags&8||Xi(this)}run(){if(!(this.flags&1))return this.fn();this.flags|=2,ai(this),Ji(this);const t=oe,n=Pe;oe=this,Pe=!0;try{return this.fn()}finally{Zi(this),oe=t,Pe=n,this.flags&=-3}}stop(){if(this.flags&1){for(let t=this.deps;t;t=t.nextDep)Vs(t);this.deps=this.depsTail=void 0,ai(this),this.onStop&&this.onStop(),this.flags&=-2}}trigger(){this.flags&64?cs.add(this):this.scheduler?this.scheduler():this.runIfDirty()}runIfDirty(){_s(this)&&this.run()}get dirty(){return _s(this)}}let Yi=0,Yt,Xt;function Xi(e,t=!1){if(e.flags|=8,t){e.next=Xt,Xt=e;return}e.next=Yt,Yt=e}function Rs(){Yi++}function Ms(){if(--Yi>0)return;if(Xt){let t=Xt;for(Xt=void 0;t;){const n=t.next;t.next=void 0,t.flags&=-9,t=n}}let e;for(;Yt;){let t=Yt;for(Yt=void 0;t;){const n=t.next;if(t.next=void 0,t.flags&=-9,t.flags&1)try{t.trigger()}catch(i){e||(e=i)}t=n}}if(e)throw e}function Ji(e){for(let t=e.deps;t;t=t.nextDep)t.version=-1,t.prevActiveLink=t.dep.activeLink,t.dep.activeLink=t}function Zi(e){let t,n=e.depsTail,i=n;for(;i;){const s=i.prevDep;i.version===-1?(i===n&&(n=s),Vs(i),wr(i)):t=i,i.dep.activeLink=i.prevActiveLink,i.prevActiveLink=void 0,i=s}e.deps=t,e.depsTail=n}function _s(e){for(let t=e.deps;t;t=t.nextDep)if(t.dep.version!==t.version||t.dep.computed&&(Qi(t.dep.computed)||t.dep.version!==t.version))return!0;return!!e._dirty}function Qi(e){if(e.flags&4&&!(e.flags&16)||(e.flags&=-17,e.globalVersion===on)||(e.globalVersion=on,!e.isSSR&&e.flags&128&&(!e.deps&&!e._dirty||!_s(e))))return;e.flags|=2;const t=e.dep,n=oe,i=Pe;oe=e,Pe=!0;try{Ji(e);const s=e.fn(e._value);(t.version===0||We(s,e._value))&&(e.flags|=128,e._value=s,t.version++)}catch(s){throw t.version++,s}finally{oe=n,Pe=i,Zi(e),e.flags&=-3}}function Vs(e,t=!1){const{dep:n,prevSub:i,nextSub:s}=e;if(i&&(i.nextSub=s,e.prevSub=void 0),s&&(s.prevSub=i,e.nextSub=void 0),n.subs===e&&(n.subs=i,!i&&n.computed)){n.computed.flags&=-5;for(let a=n.computed.deps;a;a=a.nextDep)Vs(a,!0)}!t&&!--n.sc&&n.map&&n.map.delete(n.key)}function wr(e){const{prevDep:t,nextDep:n}=e;t&&(t.nextDep=n,e.prevDep=void 0),n&&(n.prevDep=t,e.nextDep=void 0)}let Pe=!0;const ea=[];function tt(){ea.push(Pe),Pe=!1}function nt(){const e=ea.pop();Pe=e===void 0?!0:e}function ai(e){const{cleanup:t}=e;if(e.cleanup=void 0,t){const n=oe;oe=void 0;try{t()}finally{oe=n}}}let on=0;class Sr{constructor(t,n){this.sub=t,this.dep=n,this.version=n.version,this.nextDep=this.prevDep=this.nextSub=this.prevSub=this.prevActiveLink=void 0}}class Gs{constructor(t){this.computed=t,this.version=0,this.activeLink=void 0,this.subs=void 0,this.map=void 0,this.key=void 0,this.sc=0,this.__v_skip=!0}track(t){if(!oe||!Pe||oe===this.computed)return;let n=this.activeLink;if(n===void 0||n.sub!==oe)n=this.activeLink=new Sr(oe,this),oe.deps?(n.prevDep=oe.depsTail,oe.depsTail.nextDep=n,oe.depsTail=n):oe.deps=oe.depsTail=n,ta(n);else if(n.version===-1&&(n.version=this.version,n.nextDep)){const i=n.nextDep;i.prevDep=n.prevDep,n.prevDep&&(n.prevDep.nextDep=i),n.prevDep=oe.depsTail,n.nextDep=void 0,oe.depsTail.nextDep=n,oe.depsTail=n,oe.deps===n&&(oe.deps=i)}return n}trigger(t){this.version++,on++,this.notify(t)}notify(t){Rs();try{for(let n=this.subs;n;n=n.prevSub)n.sub.notify()&&n.sub.dep.notify()}finally{Ms()}}}function ta(e){if(e.dep.sc++,e.sub.flags&4){const t=e.dep.computed;if(t&&!e.dep.subs){t.flags|=20;for(let i=t.deps;i;i=i.nextDep)ta(i)}const n=e.dep.subs;n!==e&&(e.prevSub=n,n&&(n.nextSub=e)),e.dep.subs=e}}const Dn=new WeakMap,bt=Symbol(""),ws=Symbol(""),ln=Symbol("");function ve(e,t,n){if(Pe&&oe){let i=Dn.get(e);i||Dn.set(e,i=new Map);let s=i.get(n);s||(i.set(n,s=new Gs),s.map=i,s.key=n),s.track()}}function Ze(e,t,n,i,s,a){const r=Dn.get(e);if(!r){on++;return}const o=l=>{l&&l.trigger()};if(Rs(),t==="clear")r.forEach(o);else{const l=O(e),d=l&&jn(n);if(l&&n==="length"){const c=Number(i);r.forEach((h,m)=>{(m==="length"||m===ln||!Le(m)&&m>=c)&&o(h)})}else switch((n!==void 0||r.has(void 0))&&o(r.get(n)),d&&o(r.get(ln)),t){case"add":l?d&&o(r.get("length")):(o(r.get(bt)),Lt(e)&&o(r.get(ws)));break;case"delete":l||(o(r.get(bt)),Lt(e)&&o(r.get(ws)));break;case"set":Lt(e)&&o(r.get(bt));break}}Ms()}function Er(e,t){const n=Dn.get(e);return n&&n.get(t)}function It(e){const t=Z(e);return t===e?t:(ve(t,"iterate",ln),Ae(e)?t:t.map(Be))}function Un(e){return ve(e=Z(e),"iterate",ln),e}function je(e,t){return st(e)?Bt(et(e)?Be(t):t):Be(t)}const xr={__proto__:null,[Symbol.iterator](){return ds(this,Symbol.iterator,e=>je(this,e))},concat(...e){return It(this).concat(...e.map(t=>O(t)?It(t):t))},entries(){return ds(this,"entries",e=>(e[1]=je(this,e[1]),e))},every(e,t){return Ye(this,"every",e,t,void 0,arguments)},filter(e,t){return Ye(this,"filter",e,t,n=>n.map(i=>je(this,i)),arguments)},find(e,t){return Ye(this,"find",e,t,n=>je(this,n),arguments)},findIndex(e,t){return Ye(this,"findIndex",e,t,void 0,arguments)},findLast(e,t){return Ye(this,"findLast",e,t,n=>je(this,n),arguments)},findLastIndex(e,t){return Ye(this,"findLastIndex",e,t,void 0,arguments)},forEach(e,t){return Ye(this,"forEach",e,t,void 0,arguments)},includes(...e){return us(this,"includes",e)},indexOf(...e){return us(this,"indexOf",e)},join(e){return It(this).join(e)},lastIndexOf(...e){return us(this,"lastIndexOf",e)},map(e,t){return Ye(this,"map",e,t,void 0,arguments)},pop(){return Ht(this,"pop")},push(...e){return Ht(this,"push",e)},reduce(e,...t){return ri(this,"reduce",e,t)},reduceRight(e,...t){return ri(this,"reduceRight",e,t)},shift(){return Ht(this,"shift")},some(e,t){return Ye(this,"some",e,t,void 0,arguments)},splice(...e){return Ht(this,"splice",e)},toReversed(){return It(this).toReversed()},toSorted(e){return It(this).toSorted(e)},toSpliced(...e){return It(this).toSpliced(...e)},unshift(...e){return Ht(this,"unshift",e)},values(){return ds(this,"values",e=>je(this,e))}};function ds(e,t,n){const i=Un(e),s=i[t]();return i!==e&&!Ae(e)&&(s._next=s.next,s.next=()=>{const a=s._next();return a.done||(a.value=n(a.value)),a}),s}const Ir=Array.prototype;function Ye(e,t,n,i,s,a){const r=Un(e),o=r!==e&&!Ae(e),l=r[t];if(l!==Ir[t]){const h=l.apply(e,a);return o?Be(h):h}let d=n;r!==e&&(o?d=function(h,m){return n.call(this,je(e,h),m,e)}:n.length>2&&(d=function(h,m){return n.call(this,h,m,e)}));const c=l.call(r,d,i);return o&&s?s(c):c}function ri(e,t,n,i){const s=Un(e),a=s!==e&&!Ae(e);let r=n,o=!1;s!==e&&(a?(o=i.length===0,r=function(d,c,h){return o&&(o=!1,d=je(e,d)),n.call(this,d,je(e,c),h,e)}):n.length>3&&(r=function(d,c,h){return n.call(this,d,c,h,e)}));const l=s[t](r,...i);return o?je(e,l):l}function us(e,t,n){const i=Z(e);ve(i,"iterate",ln);const s=i[t](...n);return(s===-1||s===!1)&&qn(n[0])?(n[0]=Z(n[0]),i[t](...n)):s}function Ht(e,t,n=[]){tt(),Rs();const i=Z(e)[t].apply(e,n);return Ms(),nt(),i}const Cr=Ns("__proto__,__v_isRef,__isVue"),na=new Set(Object.getOwnPropertyNames(Symbol).filter(e=>e!=="arguments"&&e!=="caller").map(e=>Symbol[e]).filter(Le));function kr(e){Le(e)||(e=String(e));const t=Z(this);return ve(t,"has",e),t.hasOwnProperty(e)}class sa{constructor(t=!1,n=!1){this._isReadonly=t,this._isShallow=n}get(t,n,i){if(n==="__v_skip")return t.__v_skip;const s=this._isReadonly,a=this._isShallow;if(n==="__v_isReactive")return!s;if(n==="__v_isReadonly")return s;if(n==="__v_isShallow")return a;if(n==="__v_raw")return i===(s?a?Mr:oa:a?ra:aa).get(t)||Object.getPrototypeOf(t)===Object.getPrototypeOf(i)?t:void 0;const r=O(t);if(!s){let l;if(r&&(l=xr[n]))return l;if(n==="hasOwnProperty")return kr}const o=Reflect.get(t,n,de(t)?t:i);if((Le(n)?na.has(n):Cr(n))||(s||ve(t,"get",n),a))return o;if(de(o)){const l=r&&jn(n)?o:o.value;return s&&te(l)?Es(l):l}return te(o)?s?Es(o):Kn(o):o}}class ia extends sa{constructor(t=!1){super(!1,t)}set(t,n,i,s){let a=t[n];const r=O(t)&&jn(n);if(!this._isShallow){const d=st(a);if(!Ae(i)&&!st(i)&&(a=Z(a),i=Z(i)),!r&&de(a)&&!de(i))return d||(a.value=i),!0}const o=r?Number(n)<t.length:ee(t,n),l=Reflect.set(t,n,i,de(t)?t:s);return t===Z(s)&&(o?We(i,a)&&Ze(t,"set",n,i):Ze(t,"add",n,i)),l}deleteProperty(t,n){const i=ee(t,n);t[n];const s=Reflect.deleteProperty(t,n);return s&&i&&Ze(t,"delete",n,void 0),s}has(t,n){const i=Reflect.has(t,n);return(!Le(n)||!na.has(n))&&ve(t,"has",n),i}ownKeys(t){return ve(t,"iterate",O(t)?"length":bt),Reflect.ownKeys(t)}}class Ar extends sa{constructor(t=!1){super(!0,t)}set(t,n){return!0}deleteProperty(t,n){return!0}}const Dr=new ia,Tr=new Ar,Lr=new ia(!0);const Ss=e=>e,wn=e=>Reflect.getPrototypeOf(e);function Pr(e,t,n){return function(...i){const s=this.__v_raw,a=Z(s),r=Lt(a),o=e==="entries"||e===Symbol.iterator&&r,l=e==="keys"&&r,d=s[e](...i),c=n?Ss:t?Bt:Be;return!t&&ve(a,"iterate",l?ws:bt),ye(Object.create(d),{next(){const{value:h,done:m}=d.next();return m?{value:h,done:m}:{value:o?[c(h[0]),c(h[1])]:c(h),done:m}}})}}function Sn(e){return function(...t){return e==="delete"?!1:e==="clear"?void 0:this}}function Nr(e,t){const n={get(s){const a=this.__v_raw,r=Z(a),o=Z(s);e||(We(s,o)&&ve(r,"get",s),ve(r,"get",o));const{has:l}=wn(r),d=t?Ss:e?Bt:Be;if(l.call(r,s))return d(a.get(s));if(l.call(r,o))return d(a.get(o));a!==r&&a.get(s)},get size(){const s=this.__v_raw;return!e&&ve(Z(s),"iterate",bt),s.size},has(s){const a=this.__v_raw,r=Z(a),o=Z(s);return e||(We(s,o)&&ve(r,"has",s),ve(r,"has",o)),s===o?a.has(s):a.has(s)||a.has(o)},forEach(s,a){const r=this,o=r.__v_raw,l=Z(o),d=t?Ss:e?Bt:Be;return!e&&ve(l,"iterate",bt),o.forEach((c,h)=>s.call(a,d(c),d(h),r))}};return ye(n,e?{add:Sn("add"),set:Sn("set"),delete:Sn("delete"),clear:Sn("clear")}:{add(s){const a=Z(this),r=wn(a),o=Z(s),l=!t&&!Ae(s)&&!st(s)?o:s;return r.has.call(a,l)||We(s,l)&&r.has.call(a,s)||We(o,l)&&r.has.call(a,o)||(a.add(l),Ze(a,"add",l,l)),this},set(s,a){!t&&!Ae(a)&&!st(a)&&(a=Z(a));const r=Z(this),{has:o,get:l}=wn(r);let d=o.call(r,s);d||(s=Z(s),d=o.call(r,s));const c=l.call(r,s);return r.set(s,a),d?We(a,c)&&Ze(r,"set",s,a):Ze(r,"add",s,a),this},delete(s){const a=Z(this),{has:r,get:o}=wn(a);let l=r.call(a,s);l||(s=Z(s),l=r.call(a,s)),o&&o.call(a,s);const d=a.delete(s);return l&&Ze(a,"delete",s,void 0),d},clear(){const s=Z(this),a=s.size!==0,r=s.clear();return a&&Ze(s,"clear",void 0,void 0),r}}),["keys","values","entries",Symbol.iterator].forEach(s=>{n[s]=Pr(s,e,t)}),n}function Hs(e,t){const n=Nr(e,t);return(i,s,a)=>s==="__v_isReactive"?!e:s==="__v_isReadonly"?e:s==="__v_raw"?i:Reflect.get(ee(n,s)&&s in i?n:i,s,a)}const Br={get:Hs(!1,!1)},Fr={get:Hs(!1,!0)},Rr={get:Hs(!0,!1)};const aa=new WeakMap,ra=new WeakMap,oa=new WeakMap,Mr=new WeakMap;function Vr(e){switch(e){case"Object":case"Array":return 1;case"Map":case"Set":case"WeakMap":case"WeakSet":return 2;default:return 0}}function Gr(e){return e.__v_skip||!Object.isExtensible(e)?0:Vr(cr(e))}function Kn(e){return st(e)?e:Os(e,!1,Dr,Br,aa)}function Hr(e){return Os(e,!1,Lr,Fr,ra)}function Es(e){return Os(e,!0,Tr,Rr,oa)}function Os(e,t,n,i,s){if(!te(e)||e.__v_raw&&!(t&&e.__v_isReactive))return e;const a=Gr(e);if(a===0)return e;const r=s.get(e);if(r)return r;const o=new Proxy(e,a===2?i:n);return s.set(e,o),o}function et(e){return st(e)?et(e.__v_raw):!!(e&&e.__v_isReactive)}function st(e){return!!(e&&e.__v_isReadonly)}function Ae(e){return!!(e&&e.__v_isShallow)}function qn(e){return e?!!e.__v_raw:!1}function Z(e){const t=e&&e.__v_raw;return t?Z(t):e}function yt(e){return!ee(e,"__v_skip")&&Object.isExtensible(e)&&Oi(e,"__v_skip",!0),e}const Be=e=>te(e)?Kn(e):e,Bt=e=>te(e)?Es(e):e;function de(e){return e?e.__v_isRef===!0:!1}function Jt(e){return Or(e,!1)}function Or(e,t){return de(e)?e:new jr(e,t)}class jr{constructor(t,n){this.dep=new Gs,this.__v_isRef=!0,this.__v_isShallow=!1,this._rawValue=n?t:Z(t),this._value=n?t:Be(t),this.__v_isShallow=n}get value(){return this.dep.track(),this._value}set value(t){const n=this._rawValue,i=this.__v_isShallow||Ae(t)||st(t);t=i?t:Z(t),We(t,n)&&(this._rawValue=t,this._value=i?t:Be(t),this.dep.trigger())}}function g(e){return de(e)?e.value:e}const $r={get:(e,t,n)=>t==="__v_raw"?e:g(Reflect.get(e,t,n)),set:(e,t,n,i)=>{const s=e[t];return de(s)&&!de(n)?(s.value=n,!0):Reflect.set(e,t,n,i)}};function la(e){return et(e)?e:new Proxy(e,$r)}function zr(e){const t=O(e)?new Array(e.length):{};for(const n in e)t[n]=Ur(e,n);return t}class Wr{constructor(t,n,i){this._object=t,this._defaultValue=i,this.__v_isRef=!0,this._value=void 0,this._key=Le(n)?n:String(n),this._raw=Z(t);let s=!0,a=t;if(!O(t)||Le(this._key)||!jn(this._key))do s=!qn(a)||Ae(a);while(s&&(a=a.__v_raw));this._shallow=s}get value(){let t=this._object[this._key];return this._shallow&&(t=g(t)),this._value=t===void 0?this._defaultValue:t}set value(t){if(this._shallow&&de(this._raw[this._key])){const n=this._object[this._key];if(de(n)){n.value=t;return}}this._object[this._key]=t}get dep(){return Er(this._raw,this._key)}}function Ur(e,t,n){return new Wr(e,t,n)}class Kr{constructor(t,n,i){this.fn=t,this.setter=n,this._value=void 0,this.dep=new Gs(this),this.__v_isRef=!0,this.deps=void 0,this.depsTail=void 0,this.flags=16,this.globalVersion=on-1,this.next=void 0,this.effect=this,this.__v_isReadonly=!n,this.isSSR=i}notify(){if(this.flags|=16,!(this.flags&8)&&oe!==this)return Xi(this,!0),!0}get value(){const t=this.dep.track();return Qi(this),t&&(t.version=this.dep.version),this._value}set value(t){this.setter&&this.setter(t)}}function qr(e,t,n=!1){let i,s;return $(e)?i=e:(i=e.get,s=e.set),new Kr(i,s,n)}const En={},Tn=new WeakMap;let mt;function Yr(e,t=!1,n=mt){if(n){let i=Tn.get(n);i||Tn.set(n,i=[]),i.push(e)}}function Xr(e,t,n=se){const{immediate:i,deep:s,once:a,scheduler:r,augmentJob:o,call:l}=n,d=N=>s?N:Ae(N)||s===!1||s===0?Qe(N,1):Qe(N);let c,h,m,E,B=!1,D=!1;if(de(e)?(h=()=>e.value,B=Ae(e)):et(e)?(h=()=>d(e),B=!0):O(e)?(D=!0,B=e.some(N=>et(N)||Ae(N)),h=()=>e.map(N=>{if(de(N))return N.value;if(et(N))return d(N);if($(N))return l?l(N,2):N()})):$(e)?t?h=l?()=>l(e,2):e:h=()=>{if(m){tt();try{m()}finally{nt()}}const N=mt;mt=c;try{return l?l(e,3,[E]):e(E)}finally{mt=N}}:h=Ue,t&&s){const N=h,Y=s===!0?1/0:s;h=()=>Qe(N(),Y)}const W=Ki(),U=()=>{c.stop(),W&&W.active&&Bs(W.effects,c)};if(a&&t){const N=t;t=(...Y)=>{N(...Y),U()}}let V=D?new Array(e.length).fill(En):En;const H=N=>{if(!(!(c.flags&1)||!c.dirty&&!N))if(t){const Y=c.run();if(s||B||(D?Y.some((y,P)=>We(y,V[P])):We(Y,V))){m&&m();const y=mt;mt=c;try{const P=[Y,V===En?void 0:D&&V[0]===En?[]:V,E];V=Y,l?l(t,3,P):t(...P)}finally{mt=y}}}else c.run()};return o&&o(H),c=new qi(h),c.scheduler=r?()=>r(H,!1):H,E=N=>Yr(N,!1,c),m=c.onStop=()=>{const N=Tn.get(c);if(N){if(l)l(N,4);else for(const Y of N)Y();Tn.delete(c)}},t?i?H(!0):V=c.run():r?r(H.bind(null,!0),!0):c.run(),U.pause=c.pause.bind(c),U.resume=c.resume.bind(c),U.stop=U,U}function Qe(e,t=1/0,n){if(t<=0||!te(e)||e.__v_skip||(n=n||new Map,(n.get(e)||0)>=t))return e;if(n.set(e,t),t--,de(e))Qe(e.value,t,n);else if(O(e))for(let i=0;i<e.length;i++)Qe(e[i],t,n);else if(Mi(e)||Lt(e))e.forEach(i=>{Qe(i,t,n)});else if(Hi(e)){for(const i in e)Qe(e[i],t,n);for(const i of Object.getOwnPropertySymbols(e))Object.prototype.propertyIsEnumerable.call(e,i)&&Qe(e[i],t,n)}return e}/**
* @vue/runtime-core v3.5.33
* (c) 2018-present Yuxi (Evan) You and Vue contributors
* @license MIT
**/function vn(e,t,n,i){try{return i?e(...i):e()}catch(s){Yn(s,t,n)}}function Ke(e,t,n,i){if($(e)){const s=vn(e,t,n,i);return s&&Vi(s)&&s.catch(a=>{Yn(a,t,n)}),s}if(O(e)){const s=[];for(let a=0;a<e.length;a++)s.push(Ke(e[a],t,n,i));return s}}function Yn(e,t,n,i=!0){const s=t?t.vnode:null,{errorHandler:a,throwUnhandledErrorInProduction:r}=t&&t.appContext.config||se;if(t){let o=t.parent;const l=t.proxy,d=`https://vuejs.org/error-reference/#runtime-${n}`;for(;o;){const c=o.ec;if(c){for(let h=0;h<c.length;h++)if(c[h](e,l,d)===!1)return}o=o.parent}if(a){tt(),vn(a,null,10,[e,l,d]),nt();return}}Jr(e,n,s,i,r)}function Jr(e,t,n,i=!0,s=!1){if(s)throw e;console.error(e)}const _e=[];let Oe=-1;const Pt=[];let lt=null,Dt=0;const ca=Promise.resolve();let Ln=null;function Et(e){const t=Ln||ca;return e?t.then(this?e.bind(this):e):t}function Zr(e){let t=Oe+1,n=_e.length;for(;t<n;){const i=t+n>>>1,s=_e[i],a=cn(s);a<e||a===e&&s.flags&2?t=i+1:n=i}return t}function js(e){if(!(e.flags&1)){const t=cn(e),n=_e[_e.length-1];!n||!(e.flags&2)&&t>=cn(n)?_e.push(e):_e.splice(Zr(t),0,e),e.flags|=1,da()}}function da(){Ln||(Ln=ca.then(ha))}function Qr(e){O(e)?Pt.push(...e):lt&&e.id===-1?lt.splice(Dt+1,0,e):e.flags&1||(Pt.push(e),e.flags|=1),da()}function oi(e,t,n=Oe+1){for(;n<_e.length;n++){const i=_e[n];if(i&&i.flags&2){if(e&&i.id!==e.uid)continue;_e.splice(n,1),n--,i.flags&4&&(i.flags&=-2),i(),i.flags&4||(i.flags&=-2)}}}function ua(e){if(Pt.length){const t=[...new Set(Pt)].sort((n,i)=>cn(n)-cn(i));if(Pt.length=0,lt){lt.push(...t);return}for(lt=t,Dt=0;Dt<lt.length;Dt++){const n=lt[Dt];n.flags&4&&(n.flags&=-2),n.flags&8||n(),n.flags&=-2}lt=null,Dt=0}}const cn=e=>e.id==null?e.flags&2?-1:1/0:e.id;function ha(e){try{for(Oe=0;Oe<_e.length;Oe++){const t=_e[Oe];t&&!(t.flags&8)&&(t.flags&4&&(t.flags&=-2),vn(t,t.i,t.i?15:14),t.flags&4||(t.flags&=-2))}}finally{for(;Oe<_e.length;Oe++){const t=_e[Oe];t&&(t.flags&=-2)}Oe=-1,_e.length=0,ua(),Ln=null,(_e.length||Pt.length)&&ha()}}let Ce=null,fa=null;function Pn(e){const t=Ce;return Ce=e,fa=e&&e.type.__scopeId||null,t}function eo(e,t=Ce,n){if(!t||e._n)return e;const i=(...s)=>{i._d&&Fn(-1);const a=Pn(t);let r;try{r=e(...s)}finally{Pn(a),i._d&&Fn(1)}return r};return i._n=!0,i._c=!0,i._d=!0,i}function dn(e,t){if(Ce===null)return e;const n=Qn(Ce),i=e.dirs||(e.dirs=[]);for(let s=0;s<t.length;s++){let[a,r,o,l=se]=t[s];a&&($(a)&&(a={mounted:a,updated:a}),a.deep&&Qe(r),i.push({dir:a,instance:n,value:r,oldValue:void 0,arg:o,modifiers:l}))}return e}function ft(e,t,n,i){const s=e.dirs,a=t&&t.dirs;for(let r=0;r<s.length;r++){const o=s[r];a&&(o.oldValue=a[r].value);let l=o.dir[i];l&&(tt(),Ke(l,n,8,[e.el,o,e,t]),nt())}}function to(e,t){if(me){let n=me.provides;const i=me.parent&&me.parent.provides;i===n&&(n=me.provides=Object.create(i)),n[e]=t}}function Zt(e,t,n=!1){const i=Ga();if(i||_t){let s=_t?_t._context.provides:i?i.parent==null||i.ce?i.vnode.appContext&&i.vnode.appContext.provides:i.parent.provides:void 0;if(s&&e in s)return s[e];if(arguments.length>1)return n&&$(t)?t.call(i&&i.proxy):t}}function no(){return!!(Ga()||_t)}const so=Symbol.for("v-scx"),io=()=>Zt(so);function Nt(e,t,n){return pa(e,t,n)}function pa(e,t,n=se){const{immediate:i,deep:s,flush:a,once:r}=n,o=ye({},n),l=t&&i||!t&&a!=="post";let d;if(hn){if(a==="sync"){const E=io();d=E.__watcherHandles||(E.__watcherHandles=[])}else if(!l){const E=()=>{};return E.stop=Ue,E.resume=Ue,E.pause=Ue,E}}const c=me;o.call=(E,B,D)=>Ke(E,c,B,D);let h=!1;a==="post"?o.scheduler=E=>{xe(E,c&&c.suspense)}:a!=="sync"&&(h=!0,o.scheduler=(E,B)=>{B?E():js(E)}),o.augmentJob=E=>{t&&(E.flags|=4),h&&(E.flags|=2,c&&(E.id=c.uid,E.i=c))};const m=Xr(e,t,o);return hn&&(d?d.push(m):l&&m()),m}function ao(e,t,n){const i=this.proxy,s=ce(e)?e.includes(".")?ga(i,e):()=>i[e]:e.bind(i,i);let a;$(t)?a=t:(a=t.handler,n=t);const r=mn(this),o=pa(s,a.bind(i),n);return r(),o}function ga(e,t){const n=t.split(".");return()=>{let i=e;for(let s=0;s<n.length&&i;s++)i=i[n[s]];return i}}const ro=Symbol("_vte"),oo=e=>e.__isTeleport,lo=Symbol("_leaveCb");function $s(e,t){e.shapeFlag&6&&e.component?(e.transition=t,$s(e.component.subTree,t)):e.shapeFlag&128?(e.ssContent.transition=t.clone(e.ssContent),e.ssFallback.transition=t.clone(e.ssFallback)):e.transition=t}function va(e){e.ids=[e.ids[0]+e.ids[2]+++"-",0,0]}function li(e,t){let n;return!!((n=Object.getOwnPropertyDescriptor(e,t))&&!n.configurable)}const Nn=new WeakMap;function Qt(e,t,n,i,s=!1){if(O(e)){e.forEach((D,W)=>Qt(D,t&&(O(t)?t[W]:t),n,i,s));return}if(en(i)&&!s){i.shapeFlag&512&&i.type.__asyncResolved&&i.component.subTree.component&&Qt(e,t,n,i.component.subTree);return}const a=i.shapeFlag&4?Qn(i.component):i.el,r=s?null:a,{i:o,r:l}=e,d=t&&t.r,c=o.refs===se?o.refs={}:o.refs,h=o.setupState,m=Z(h),E=h===se?Ri:D=>li(c,D)?!1:ee(m,D),B=(D,W)=>!(W&&li(c,W));if(d!=null&&d!==l){if(ci(t),ce(d))c[d]=null,E(d)&&(h[d]=null);else if(de(d)){const D=t;B(d,D.k)&&(d.value=null),D.k&&(c[D.k]=null)}}if($(l))vn(l,o,12,[r,c]);else{const D=ce(l),W=de(l);if(D||W){const U=()=>{if(e.f){const V=D?E(l)?h[l]:c[l]:B()||!e.k?l.value:c[e.k];if(s)O(V)&&Bs(V,a);else if(O(V))V.includes(a)||V.push(a);else if(D)c[l]=[a],E(l)&&(h[l]=c[l]);else{const H=[a];B(l,e.k)&&(l.value=H),e.k&&(c[e.k]=H)}}else D?(c[l]=r,E(l)&&(h[l]=r)):W&&(B(l,e.k)&&(l.value=r),e.k&&(c[e.k]=r))};if(r){const V=()=>{U(),Nn.delete(e)};V.id=-1,Nn.set(e,V),xe(V,n)}else ci(e),U()}}}function ci(e){const t=Nn.get(e);t&&(t.flags|=8,Nn.delete(e))}Wn().requestIdleCallback;Wn().cancelIdleCallback;const en=e=>!!e.type.__asyncLoader,ma=e=>e.type.__isKeepAlive;function co(e,t){ya(e,"a",t)}function uo(e,t){ya(e,"da",t)}function ya(e,t,n=me){const i=e.__wdc||(e.__wdc=()=>{let s=n;for(;s;){if(s.isDeactivated)return;s=s.parent}return e()});if(Xn(t,i,n),n){let s=n.parent;for(;s&&s.parent;)ma(s.parent.vnode)&&ho(i,t,n,s),s=s.parent}}function ho(e,t,n,i){const s=Xn(t,e,i,!0);Ws(()=>{Bs(i[t],s)},n)}function Xn(e,t,n=me,i=!1){if(n){const s=n[e]||(n[e]=[]),a=t.__weh||(t.__weh=(...r)=>{tt();const o=mn(n),l=Ke(t,n,e,r);return o(),nt(),l});return i?s.unshift(a):s.push(a),a}}const at=e=>(t,n=me)=>{(!hn||e==="sp")&&Xn(e,(...i)=>t(...i),n)},fo=at("bm"),zs=at("m"),po=at("bu"),go=at("u"),ba=at("bum"),Ws=at("um"),vo=at("sp"),mo=at("rtg"),yo=at("rtc");function bo(e,t=me){Xn("ec",e,t)}const _o="components";function wo(e,t){return Eo(_o,e,!0,t)||e}const So=Symbol.for("v-ndc");function Eo(e,t,n=!0,i=!1){const s=Ce||me;if(s){const a=s.type;{const o=ll(a,!1);if(o&&(o===t||o===we(t)||o===zn(we(t))))return a}const r=di(s[e]||a[e],t)||di(s.appContext[e],t);return!r&&i?a:r}}function di(e,t){return e&&(e[t]||e[we(t)]||e[zn(we(t))])}function Se(e,t,n,i){let s;const a=n,r=O(e);if(r||ce(e)){const o=r&&et(e);let l=!1,d=!1;o&&(l=!Ae(e),d=st(e),e=Un(e)),s=new Array(e.length);for(let c=0,h=e.length;c<h;c++)s[c]=t(l?d?Bt(Be(e[c])):Be(e[c]):e[c],c,void 0,a)}else if(typeof e=="number"){s=new Array(e);for(let o=0;o<e;o++)s[o]=t(o+1,o,void 0,a)}else if(te(e))if(e[Symbol.iterator])s=Array.from(e,(o,l)=>t(o,l,void 0,a));else{const o=Object.keys(e);s=new Array(o.length);for(let l=0,d=o.length;l<d;l++){const c=o[l];s[l]=t(e[c],c,l,a)}}else s=[];return s}const xs=e=>e?Ha(e)?Qn(e):xs(e.parent):null,tn=ye(Object.create(null),{$:e=>e,$el:e=>e.vnode.el,$data:e=>e.data,$props:e=>e.props,$attrs:e=>e.attrs,$slots:e=>e.slots,$refs:e=>e.refs,$parent:e=>xs(e.parent),$root:e=>xs(e.root),$host:e=>e.ce,$emit:e=>e.emit,$options:e=>wa(e),$forceUpdate:e=>e.f||(e.f=()=>{js(e.update)}),$nextTick:e=>e.n||(e.n=Et.bind(e.proxy)),$watch:e=>ao.bind(e)}),hs=(e,t)=>e!==se&&!e.__isScriptSetup&&ee(e,t),xo={get({_:e},t){if(t==="__v_skip")return!0;const{ctx:n,setupState:i,data:s,props:a,accessCache:r,type:o,appContext:l}=e;if(t[0]!=="$"){const m=r[t];if(m!==void 0)switch(m){case 1:return i[t];case 2:return s[t];case 4:return n[t];case 3:return a[t]}else{if(hs(i,t))return r[t]=1,i[t];if(s!==se&&ee(s,t))return r[t]=2,s[t];if(ee(a,t))return r[t]=3,a[t];if(n!==se&&ee(n,t))return r[t]=4,n[t];Is&&(r[t]=0)}}const d=tn[t];let c,h;if(d)return t==="$attrs"&&ve(e.attrs,"get",""),d(e);if((c=o.__cssModules)&&(c=c[t]))return c;if(n!==se&&ee(n,t))return r[t]=4,n[t];if(h=l.config.globalProperties,ee(h,t))return h[t]},set({_:e},t,n){const{data:i,setupState:s,ctx:a}=e;return hs(s,t)?(s[t]=n,!0):i!==se&&ee(i,t)?(i[t]=n,!0):ee(e.props,t)||t[0]==="$"&&t.slice(1)in e?!1:(a[t]=n,!0)},has({_:{data:e,setupState:t,accessCache:n,ctx:i,appContext:s,props:a,type:r}},o){let l;return!!(n[o]||e!==se&&o[0]!=="$"&&ee(e,o)||hs(t,o)||ee(a,o)||ee(i,o)||ee(tn,o)||ee(s.config.globalProperties,o)||(l=r.__cssModules)&&l[o])},defineProperty(e,t,n){return n.get!=null?e._.accessCache[t]=0:ee(n,"value")&&this.set(e,t,n.value,null),Reflect.defineProperty(e,t,n)}};function ui(e){return O(e)?e.reduce((t,n)=>(t[n]=null,t),{}):e}let Is=!0;function Io(e){const t=wa(e),n=e.proxy,i=e.ctx;Is=!1,t.beforeCreate&&hi(t.beforeCreate,e,"bc");const{data:s,computed:a,methods:r,watch:o,provide:l,inject:d,created:c,beforeMount:h,mounted:m,beforeUpdate:E,updated:B,activated:D,deactivated:W,beforeDestroy:U,beforeUnmount:V,destroyed:H,unmounted:N,render:Y,renderTracked:y,renderTriggered:P,errorCaptured:G,serverPrefetch:z,expose:le,inheritAttrs:De,components:Fe,directives:rt,filters:Rt}=t;if(d&&Co(d,i,null),r)for(const K in r){const ae=r[K];$(ae)&&(i[K]=ae.bind(n))}if(s){const K=s.call(n,n);te(K)&&(e.data=Kn(K))}if(Is=!0,a)for(const K in a){const ae=a[K],ut=$(ae)?ae.bind(n,n):$(ae.get)?ae.get.bind(n,n):Ue,bn=!$(ae)&&$(ae.set)?ae.set.bind(n):Ue,ht=Ne({get:ut,set:bn});Object.defineProperty(i,K,{enumerable:!0,configurable:!0,get:()=>ht.value,set:Re=>ht.value=Re})}if(o)for(const K in o)_a(o[K],i,n,K);if(l){const K=$(l)?l.call(n):l;Reflect.ownKeys(K).forEach(ae=>{to(ae,K[ae])})}c&&hi(c,e,"c");function ue(K,ae){O(ae)?ae.forEach(ut=>K(ut.bind(n))):ae&&K(ae.bind(n))}if(ue(fo,h),ue(zs,m),ue(po,E),ue(go,B),ue(co,D),ue(uo,W),ue(bo,G),ue(yo,y),ue(mo,P),ue(ba,V),ue(Ws,N),ue(vo,z),O(le))if(le.length){const K=e.exposed||(e.exposed={});le.forEach(ae=>{Object.defineProperty(K,ae,{get:()=>n[ae],set:ut=>n[ae]=ut,enumerable:!0})})}else e.exposed||(e.exposed={});Y&&e.render===Ue&&(e.render=Y),De!=null&&(e.inheritAttrs=De),Fe&&(e.components=Fe),rt&&(e.directives=rt),z&&va(e)}function Co(e,t,n=Ue){O(e)&&(e=Cs(e));for(const i in e){const s=e[i];let a;te(s)?"default"in s?a=Zt(s.from||i,s.default,!0):a=Zt(s.from||i):a=Zt(s),de(a)?Object.defineProperty(t,i,{enumerable:!0,configurable:!0,get:()=>a.value,set:r=>a.value=r}):t[i]=a}}function hi(e,t,n){Ke(O(e)?e.map(i=>i.bind(t.proxy)):e.bind(t.proxy),t,n)}function _a(e,t,n,i){let s=i.includes(".")?ga(n,i):()=>n[i];if(ce(e)){const a=t[e];$(a)&&Nt(s,a)}else if($(e))Nt(s,e.bind(n));else if(te(e))if(O(e))e.forEach(a=>_a(a,t,n,i));else{const a=$(e.handler)?e.handler.bind(n):t[e.handler];$(a)&&Nt(s,a,e)}}function wa(e){const t=e.type,{mixins:n,extends:i}=t,{mixins:s,optionsCache:a,config:{optionMergeStrategies:r}}=e.appContext,o=a.get(t);let l;return o?l=o:!s.length&&!n&&!i?l=t:(l={},s.length&&s.forEach(d=>Bn(l,d,r,!0)),Bn(l,t,r)),te(t)&&a.set(t,l),l}function Bn(e,t,n,i=!1){const{mixins:s,extends:a}=t;a&&Bn(e,a,n,!0),s&&s.forEach(r=>Bn(e,r,n,!0));for(const r in t)if(!(i&&r==="expose")){const o=ko[r]||n&&n[r];e[r]=o?o(e[r],t[r]):t[r]}return e}const ko={data:fi,props:pi,emits:pi,methods:Ut,computed:Ut,beforeCreate:be,created:be,beforeMount:be,mounted:be,beforeUpdate:be,updated:be,beforeDestroy:be,beforeUnmount:be,destroyed:be,unmounted:be,activated:be,deactivated:be,errorCaptured:be,serverPrefetch:be,components:Ut,directives:Ut,watch:Do,provide:fi,inject:Ao};function fi(e,t){return t?e?function(){return ye($(e)?e.call(this,this):e,$(t)?t.call(this,this):t)}:t:e}function Ao(e,t){return Ut(Cs(e),Cs(t))}function Cs(e){if(O(e)){const t={};for(let n=0;n<e.length;n++)t[e[n]]=e[n];return t}return e}function be(e,t){return e?[...new Set([].concat(e,t))]:t}function Ut(e,t){return e?ye(Object.create(null),e,t):t}function pi(e,t){return e?O(e)&&O(t)?[...new Set([...e,...t])]:ye(Object.create(null),ui(e),ui(t??{})):t}function Do(e,t){if(!e)return t;if(!t)return e;const n=ye(Object.create(null),e);for(const i in t)n[i]=be(e[i],t[i]);return n}function Sa(){return{app:null,config:{isNativeTag:Ri,performance:!1,globalProperties:{},optionMergeStrategies:{},errorHandler:void 0,warnHandler:void 0,compilerOptions:{}},mixins:[],components:{},directives:{},provides:Object.create(null),optionsCache:new WeakMap,propsCache:new WeakMap,emitsCache:new WeakMap}}let To=0;function Lo(e,t){return function(i,s=null){$(i)||(i=ye({},i)),s!=null&&!te(s)&&(s=null);const a=Sa(),r=new WeakSet,o=[];let l=!1;const d=a.app={_uid:To++,_component:i,_props:s,_container:null,_context:a,_instance:null,version:ul,get config(){return a.config},set config(c){},use(c,...h){return r.has(c)||(c&&$(c.install)?(r.add(c),c.install(d,...h)):$(c)&&(r.add(c),c(d,...h))),d},mixin(c){return a.mixins.includes(c)||a.mixins.push(c),d},component(c,h){return h?(a.components[c]=h,d):a.components[c]},directive(c,h){return h?(a.directives[c]=h,d):a.directives[c]},mount(c,h,m){if(!l){const E=d._ceVNode||q(i,s);return E.appContext=a,m===!0?m="svg":m===!1&&(m=void 0),e(E,c,m),l=!0,d._container=c,c.__vue_app__=d,Qn(E.component)}},onUnmount(c){o.push(c)},unmount(){l&&(Ke(o,d._instance,16),e(null,d._container),delete d._container.__vue_app__)},provide(c,h){return a.provides[c]=h,d},runWithContext(c){const h=_t;_t=d;try{return c()}finally{_t=h}}};return d}}let _t=null;const Po=(e,t)=>t==="modelValue"||t==="model-value"?e.modelModifiers:e[`${t}Modifiers`]||e[`${we(t)}Modifiers`]||e[`${St(t)}Modifiers`];function No(e,t,...n){if(e.isUnmounted)return;const i=e.vnode.props||se;let s=n;const a=t.startsWith("update:"),r=a&&Po(i,t.slice(7));r&&(r.trim&&(s=n.map(c=>ce(c)?c.trim():c)),r.number&&(s=n.map(hr)));let o,l=i[o=rs(t)]||i[o=rs(we(t))];!l&&a&&(l=i[o=rs(St(t))]),l&&Ke(l,e,6,s);const d=i[o+"Once"];if(d){if(!e.emitted)e.emitted={};else if(e.emitted[o])return;e.emitted[o]=!0,Ke(d,e,6,s)}}const Bo=new WeakMap;function Ea(e,t,n=!1){const i=n?Bo:t.emitsCache,s=i.get(e);if(s!==void 0)return s;const a=e.emits;let r={},o=!1;if(!$(e)){const l=d=>{const c=Ea(d,t,!0);c&&(o=!0,ye(r,c))};!n&&t.mixins.length&&t.mixins.forEach(l),e.extends&&l(e.extends),e.mixins&&e.mixins.forEach(l)}return!a&&!o?(te(e)&&i.set(e,null),null):(O(a)?a.forEach(l=>r[l]=null):ye(r,a),te(e)&&i.set(e,r),r)}function Jn(e,t){return!e||!Hn(t)?!1:(t=t.slice(2).replace(/Once$/,""),ee(e,t[0].toLowerCase()+t.slice(1))||ee(e,St(t))||ee(e,t))}function gi(e){const{type:t,vnode:n,proxy:i,withProxy:s,propsOptions:[a],slots:r,attrs:o,emit:l,render:d,renderCache:c,props:h,data:m,setupState:E,ctx:B,inheritAttrs:D}=e,W=Pn(e);let U,V;try{if(n.shapeFlag&4){const N=s||i,Y=N;U=$e(d.call(Y,N,c,h,E,m,B)),V=o}else{const N=t;U=$e(N.length>1?N(h,{attrs:o,slots:r,emit:l}):N(h,null)),V=t.props?o:Fo(o)}}catch(N){nn.length=0,Yn(N,e,1),U=q(ct)}let H=U;if(V&&D!==!1){const N=Object.keys(V),{shapeFlag:Y}=H;N.length&&Y&7&&(a&&N.some(On)&&(V=Ro(V,a)),H=Ft(H,V,!1,!0))}return n.dirs&&(H=Ft(H,null,!1,!0),H.dirs=H.dirs?H.dirs.concat(n.dirs):n.dirs),n.transition&&$s(H,n.transition),U=H,Pn(W),U}const Fo=e=>{let t;for(const n in e)(n==="class"||n==="style"||Hn(n))&&((t||(t={}))[n]=e[n]);return t},Ro=(e,t)=>{const n={};for(const i in e)(!On(i)||!(i.slice(9)in t))&&(n[i]=e[i]);return n};function Mo(e,t,n){const{props:i,children:s,component:a}=e,{props:r,children:o,patchFlag:l}=t,d=a.emitsOptions;if(t.dirs||t.transition)return!0;if(n&&l>=0){if(l&1024)return!0;if(l&16)return i?vi(i,r,d):!!r;if(l&8){const c=t.dynamicProps;for(let h=0;h<c.length;h++){const m=c[h];if(xa(r,i,m)&&!Jn(d,m))return!0}}}else return(s||o)&&(!o||!o.$stable)?!0:i===r?!1:i?r?vi(i,r,d):!0:!!r;return!1}function vi(e,t,n){const i=Object.keys(t);if(i.length!==Object.keys(e).length)return!0;for(let s=0;s<i.length;s++){const a=i[s];if(xa(t,e,a)&&!Jn(n,a))return!0}return!1}function xa(e,t,n){const i=e[n],s=t[n];return n==="style"&&te(i)&&te(s)?!Fs(i,s):i!==s}function Vo({vnode:e,parent:t,suspense:n},i){for(;t;){const s=t.subTree;if(s.suspense&&s.suspense.activeBranch===e&&(s.suspense.vnode.el=s.el=i,e=s),s===e)(e=t.vnode).el=i,t=t.parent;else break}n&&n.activeBranch===e&&(n.vnode.el=i)}const Ia={},Ca=()=>Object.create(Ia),ka=e=>Object.getPrototypeOf(e)===Ia;function Go(e,t,n,i=!1){const s={},a=Ca();e.propsDefaults=Object.create(null),Aa(e,t,s,a);for(const r in e.propsOptions[0])r in s||(s[r]=void 0);n?e.props=i?s:Hr(s):e.type.props?e.props=s:e.props=a,e.attrs=a}function Ho(e,t,n,i){const{props:s,attrs:a,vnode:{patchFlag:r}}=e,o=Z(s),[l]=e.propsOptions;let d=!1;if((i||r>0)&&!(r&16)){if(r&8){const c=e.vnode.dynamicProps;for(let h=0;h<c.length;h++){let m=c[h];if(Jn(e.emitsOptions,m))continue;const E=t[m];if(l)if(ee(a,m))E!==a[m]&&(a[m]=E,d=!0);else{const B=we(m);s[B]=ks(l,o,B,E,e,!1)}else E!==a[m]&&(a[m]=E,d=!0)}}}else{Aa(e,t,s,a)&&(d=!0);let c;for(const h in o)(!t||!ee(t,h)&&((c=St(h))===h||!ee(t,c)))&&(l?n&&(n[h]!==void 0||n[c]!==void 0)&&(s[h]=ks(l,o,h,void 0,e,!0)):delete s[h]);if(a!==o)for(const h in a)(!t||!ee(t,h))&&(delete a[h],d=!0)}d&&Ze(e.attrs,"set","")}function Aa(e,t,n,i){const[s,a]=e.propsOptions;let r=!1,o;if(t)for(let l in t){if(qt(l))continue;const d=t[l];let c;s&&ee(s,c=we(l))?!a||!a.includes(c)?n[c]=d:(o||(o={}))[c]=d:Jn(e.emitsOptions,l)||(!(l in i)||d!==i[l])&&(i[l]=d,r=!0)}if(a){const l=Z(n),d=o||se;for(let c=0;c<a.length;c++){const h=a[c];n[h]=ks(s,l,h,d[h],e,!ee(d,h))}}return r}function ks(e,t,n,i,s,a){const r=e[n];if(r!=null){const o=ee(r,"default");if(o&&i===void 0){const l=r.default;if(r.type!==Function&&!r.skipFactory&&$(l)){const{propsDefaults:d}=s;if(n in d)i=d[n];else{const c=mn(s);i=d[n]=l.call(null,t),c()}}else i=l;s.ce&&s.ce._setProp(n,i)}r[0]&&(a&&!o?i=!1:r[1]&&(i===""||i===St(n))&&(i=!0))}return i}const Oo=new WeakMap;function Da(e,t,n=!1){const i=n?Oo:t.propsCache,s=i.get(e);if(s)return s;const a=e.props,r={},o=[];let l=!1;if(!$(e)){const c=h=>{l=!0;const[m,E]=Da(h,t,!0);ye(r,m),E&&o.push(...E)};!n&&t.mixins.length&&t.mixins.forEach(c),e.extends&&c(e.extends),e.mixins&&e.mixins.forEach(c)}if(!a&&!l)return te(e)&&i.set(e,Tt),Tt;if(O(a))for(let c=0;c<a.length;c++){const h=we(a[c]);mi(h)&&(r[h]=se)}else if(a)for(const c in a){const h=we(c);if(mi(h)){const m=a[c],E=r[h]=O(m)||$(m)?{type:m}:ye({},m),B=E.type;let D=!1,W=!0;if(O(B))for(let U=0;U<B.length;++U){const V=B[U],H=$(V)&&V.name;if(H==="Boolean"){D=!0;break}else H==="String"&&(W=!1)}else D=$(B)&&B.name==="Boolean";E[0]=D,E[1]=W,(D||ee(E,"default"))&&o.push(h)}}const d=[r,o];return te(e)&&i.set(e,d),d}function mi(e){return e[0]!=="$"&&!qt(e)}const Us=e=>e==="_"||e==="_ctx"||e==="$stable",Ks=e=>O(e)?e.map($e):[$e(e)],jo=(e,t,n)=>{if(t._n)return t;const i=eo((...s)=>Ks(t(...s)),n);return i._c=!1,i},Ta=(e,t,n)=>{const i=e._ctx;for(const s in e){if(Us(s))continue;const a=e[s];if($(a))t[s]=jo(s,a,i);else if(a!=null){const r=Ks(a);t[s]=()=>r}}},La=(e,t)=>{const n=Ks(t);e.slots.default=()=>n},Pa=(e,t,n)=>{for(const i in t)(n||!Us(i))&&(e[i]=t[i])},$o=(e,t,n)=>{const i=e.slots=Ca();if(e.vnode.shapeFlag&32){const s=t._;s?(Pa(i,t,n),n&&Oi(i,"_",s,!0)):Ta(t,i)}else t&&La(e,t)},zo=(e,t,n)=>{const{vnode:i,slots:s}=e;let a=!0,r=se;if(i.shapeFlag&32){const o=t._;o?n&&o===1?a=!1:Pa(s,t,n):(a=!t.$stable,Ta(t,s)),r=t}else t&&(La(e,t),r={default:1});if(a)for(const o in s)!Us(o)&&r[o]==null&&delete s[o]},xe=Yo;function Wo(e){return Uo(e)}function Uo(e,t){const n=Wn();n.__VUE__=!0;const{insert:i,remove:s,patchProp:a,createElement:r,createText:o,createComment:l,setText:d,setElementText:c,parentNode:h,nextSibling:m,setScopeId:E=Ue,insertStaticContent:B}=e,D=(u,p,v,S=null,b=null,_=null,C=void 0,I=null,x=!!p.dynamicChildren)=>{if(u===p)return;u&&!Ot(u,p)&&(S=_n(u),Re(u,b,_,!0),u=null),p.patchFlag===-2&&(x=!1,p.dynamicChildren=null);const{type:w,ref:M,shapeFlag:A}=p;switch(w){case Zn:W(u,p,v,S);break;case ct:U(u,p,v,S);break;case kn:u==null&&V(p,v,S,C);break;case Q:Fe(u,p,v,S,b,_,C,I,x);break;default:A&1?Y(u,p,v,S,b,_,C,I,x):A&6?rt(u,p,v,S,b,_,C,I,x):(A&64||A&128)&&w.process(u,p,v,S,b,_,C,I,x,Vt)}M!=null&&b?Qt(M,u&&u.ref,_,p||u,!p):M==null&&u&&u.ref!=null&&Qt(u.ref,null,_,u,!0)},W=(u,p,v,S)=>{if(u==null)i(p.el=o(p.children),v,S);else{const b=p.el=u.el;p.children!==u.children&&d(b,p.children)}},U=(u,p,v,S)=>{u==null?i(p.el=l(p.children||""),v,S):p.el=u.el},V=(u,p,v,S)=>{[u.el,u.anchor]=B(u.children,p,v,S,u.el,u.anchor)},H=({el:u,anchor:p},v,S)=>{let b;for(;u&&u!==p;)b=m(u),i(u,v,S),u=b;i(p,v,S)},N=({el:u,anchor:p})=>{let v;for(;u&&u!==p;)v=m(u),s(u),u=v;s(p)},Y=(u,p,v,S,b,_,C,I,x)=>{if(p.type==="svg"?C="svg":p.type==="math"&&(C="mathml"),u==null)y(p,v,S,b,_,C,I,x);else{const w=u.el&&u.el._isVueCE?u.el:null;try{w&&w._beginPatch(),z(u,p,b,_,C,I,x)}finally{w&&w._endPatch()}}},y=(u,p,v,S,b,_,C,I)=>{let x,w;const{props:M,shapeFlag:A,transition:F,dirs:j}=u;if(x=u.el=r(u.type,_,M&&M.is,M),A&8?c(x,u.children):A&16&&G(u.children,x,null,S,b,fs(u,_),C,I),j&&ft(u,null,S,"created"),P(x,u,u.scopeId,C,S),M){for(const ne in M)ne!=="value"&&!qt(ne)&&a(x,ne,null,M[ne],_,S);"value"in M&&a(x,"value",null,M.value,_),(w=M.onVnodeBeforeMount)&&He(w,S,u)}j&&ft(u,null,S,"beforeMount");const J=Ko(b,F);J&&F.beforeEnter(x),i(x,p,v),((w=M&&M.onVnodeMounted)||J||j)&&xe(()=>{try{w&&He(w,S,u),J&&F.enter(x),j&&ft(u,null,S,"mounted")}finally{}},b)},P=(u,p,v,S,b)=>{if(v&&E(u,v),S)for(let _=0;_<S.length;_++)E(u,S[_]);if(b){let _=b.subTree;if(p===_||Ra(_.type)&&(_.ssContent===p||_.ssFallback===p)){const C=b.vnode;P(u,C,C.scopeId,C.slotScopeIds,b.parent)}}},G=(u,p,v,S,b,_,C,I,x=0)=>{for(let w=x;w<u.length;w++){const M=u[w]=I?Je(u[w]):$e(u[w]);D(null,M,p,v,S,b,_,C,I)}},z=(u,p,v,S,b,_,C)=>{const I=p.el=u.el;let{patchFlag:x,dynamicChildren:w,dirs:M}=p;x|=u.patchFlag&16;const A=u.props||se,F=p.props||se;let j;if(v&&pt(v,!1),(j=F.onVnodeBeforeUpdate)&&He(j,v,p,u),M&&ft(p,u,v,"beforeUpdate"),v&&pt(v,!0),(A.innerHTML&&F.innerHTML==null||A.textContent&&F.textContent==null)&&c(I,""),w?le(u.dynamicChildren,w,I,v,S,fs(p,b),_):C||ae(u,p,I,null,v,S,fs(p,b),_,!1),x>0){if(x&16)De(I,A,F,v,b);else if(x&2&&A.class!==F.class&&a(I,"class",null,F.class,b),x&4&&a(I,"style",A.style,F.style,b),x&8){const J=p.dynamicProps;for(let ne=0;ne<J.length;ne++){const re=J[ne],he=A[re],pe=F[re];(pe!==he||re==="value")&&a(I,re,he,pe,b,v)}}x&1&&u.children!==p.children&&c(I,p.children)}else!C&&w==null&&De(I,A,F,v,b);((j=F.onVnodeUpdated)||M)&&xe(()=>{j&&He(j,v,p,u),M&&ft(p,u,v,"updated")},S)},le=(u,p,v,S,b,_,C)=>{for(let I=0;I<p.length;I++){const x=u[I],w=p[I],M=x.el&&(x.type===Q||!Ot(x,w)||x.shapeFlag&198)?h(x.el):v;D(x,w,M,null,S,b,_,C,!0)}},De=(u,p,v,S,b)=>{if(p!==v){if(p!==se)for(const _ in p)!qt(_)&&!(_ in v)&&a(u,_,p[_],null,b,S);for(const _ in v){if(qt(_))continue;const C=v[_],I=p[_];C!==I&&_!=="value"&&a(u,_,I,C,b,S)}"value"in v&&a(u,"value",p.value,v.value,b)}},Fe=(u,p,v,S,b,_,C,I,x)=>{const w=p.el=u?u.el:o(""),M=p.anchor=u?u.anchor:o("");let{patchFlag:A,dynamicChildren:F,slotScopeIds:j}=p;j&&(I=I?I.concat(j):j),u==null?(i(w,v,S),i(M,v,S),G(p.children||[],v,M,b,_,C,I,x)):A>0&&A&64&&F&&u.dynamicChildren&&u.dynamicChildren.length===F.length?(le(u.dynamicChildren,F,v,b,_,C,I),(p.key!=null||b&&p===b.subTree)&&Na(u,p,!0)):ae(u,p,v,M,b,_,C,I,x)},rt=(u,p,v,S,b,_,C,I,x)=>{p.slotScopeIds=I,u==null?p.shapeFlag&512?b.ctx.activate(p,v,S,C,x):Rt(p,v,S,b,_,C,x):yn(u,p,x)},Rt=(u,p,v,S,b,_,C)=>{const I=u.component=sl(u,S,b);if(ma(u)&&(I.ctx.renderer=Vt),il(I,!1,C),I.asyncDep){if(b&&b.registerDep(I,ue,C),!u.el){const x=I.subTree=q(ct);U(null,x,p,v),u.placeholder=x.el}}else ue(I,u,p,v,b,_,C)},yn=(u,p,v)=>{const S=p.component=u.component;if(Mo(u,p,v))if(S.asyncDep&&!S.asyncResolved){K(S,p,v);return}else S.next=p,S.update();else p.el=u.el,S.vnode=p},ue=(u,p,v,S,b,_,C)=>{const I=()=>{if(u.isMounted){let{next:A,bu:F,u:j,parent:J,vnode:ne}=u;{const Ve=Ba(u);if(Ve){A&&(A.el=ne.el,K(u,A,C)),Ve.asyncDep.then(()=>{xe(()=>{u.isUnmounted||w()},b)});return}}let re=A,he;pt(u,!1),A?(A.el=ne.el,K(u,A,C)):A=ne,F&&os(F),(he=A.props&&A.props.onVnodeBeforeUpdate)&&He(he,J,A,ne),pt(u,!0);const pe=gi(u),Me=u.subTree;u.subTree=pe,D(Me,pe,h(Me.el),_n(Me),u,b,_),A.el=pe.el,re===null&&Vo(u,pe.el),j&&xe(j,b),(he=A.props&&A.props.onVnodeUpdated)&&xe(()=>He(he,J,A,ne),b)}else{let A;const{el:F,props:j}=p,{bm:J,m:ne,parent:re,root:he,type:pe}=u,Me=en(p);pt(u,!1),J&&os(J),!Me&&(A=j&&j.onVnodeBeforeMount)&&He(A,re,p),pt(u,!0);{he.ce&&he.ce._hasShadowRoot()&&he.ce._injectChildStyle(pe,u.parent?u.parent.type:void 0);const Ve=u.subTree=gi(u);D(null,Ve,v,S,u,b,_),p.el=Ve.el}if(ne&&xe(ne,b),!Me&&(A=j&&j.onVnodeMounted)){const Ve=p;xe(()=>He(A,re,Ve),b)}(p.shapeFlag&256||re&&en(re.vnode)&&re.vnode.shapeFlag&256)&&u.a&&xe(u.a,b),u.isMounted=!0,p=v=S=null}};u.scope.on();const x=u.effect=new qi(I);u.scope.off();const w=u.update=x.run.bind(x),M=u.job=x.runIfDirty.bind(x);M.i=u,M.id=u.uid,x.scheduler=()=>js(M),pt(u,!0),w()},K=(u,p,v)=>{p.component=u;const S=u.vnode.props;u.vnode=p,u.next=null,Ho(u,p.props,S,v),zo(u,p.children,v),tt(),oi(u),nt()},ae=(u,p,v,S,b,_,C,I,x=!1)=>{const w=u&&u.children,M=u?u.shapeFlag:0,A=p.children,{patchFlag:F,shapeFlag:j}=p;if(F>0){if(F&128){bn(w,A,v,S,b,_,C,I,x);return}else if(F&256){ut(w,A,v,S,b,_,C,I,x);return}}j&8?(M&16&&Mt(w,b,_),A!==w&&c(v,A)):M&16?j&16?bn(w,A,v,S,b,_,C,I,x):Mt(w,b,_,!0):(M&8&&c(v,""),j&16&&G(A,v,S,b,_,C,I,x))},ut=(u,p,v,S,b,_,C,I,x)=>{u=u||Tt,p=p||Tt;const w=u.length,M=p.length,A=Math.min(w,M);let F;for(F=0;F<A;F++){const j=p[F]=x?Je(p[F]):$e(p[F]);D(u[F],j,v,null,b,_,C,I,x)}w>M?Mt(u,b,_,!0,!1,A):G(p,v,S,b,_,C,I,x,A)},bn=(u,p,v,S,b,_,C,I,x)=>{let w=0;const M=p.length;let A=u.length-1,F=M-1;for(;w<=A&&w<=F;){const j=u[w],J=p[w]=x?Je(p[w]):$e(p[w]);if(Ot(j,J))D(j,J,v,null,b,_,C,I,x);else break;w++}for(;w<=A&&w<=F;){const j=u[A],J=p[F]=x?Je(p[F]):$e(p[F]);if(Ot(j,J))D(j,J,v,null,b,_,C,I,x);else break;A--,F--}if(w>A){if(w<=F){const j=F+1,J=j<M?p[j].el:S;for(;w<=F;)D(null,p[w]=x?Je(p[w]):$e(p[w]),v,J,b,_,C,I,x),w++}}else if(w>F)for(;w<=A;)Re(u[w],b,_,!0),w++;else{const j=w,J=w,ne=new Map;for(w=J;w<=F;w++){const Ie=p[w]=x?Je(p[w]):$e(p[w]);Ie.key!=null&&ne.set(Ie.key,w)}let re,he=0;const pe=F-J+1;let Me=!1,Ve=0;const Gt=new Array(pe);for(w=0;w<pe;w++)Gt[w]=0;for(w=j;w<=A;w++){const Ie=u[w];if(he>=pe){Re(Ie,b,_,!0);continue}let Ge;if(Ie.key!=null)Ge=ne.get(Ie.key);else for(re=J;re<=F;re++)if(Gt[re-J]===0&&Ot(Ie,p[re])){Ge=re;break}Ge===void 0?Re(Ie,b,_,!0):(Gt[Ge-J]=w+1,Ge>=Ve?Ve=Ge:Me=!0,D(Ie,p[Ge],v,null,b,_,C,I,x),he++)}const ei=Me?qo(Gt):Tt;for(re=ei.length-1,w=pe-1;w>=0;w--){const Ie=J+w,Ge=p[Ie],ti=p[Ie+1],ni=Ie+1<M?ti.el||Fa(ti):S;Gt[w]===0?D(null,Ge,v,ni,b,_,C,I,x):Me&&(re<0||w!==ei[re]?ht(Ge,v,ni,2):re--)}}},ht=(u,p,v,S,b=null)=>{const{el:_,type:C,transition:I,children:x,shapeFlag:w}=u;if(w&6){ht(u.component.subTree,p,v,S);return}if(w&128){u.suspense.move(p,v,S);return}if(w&64){C.move(u,p,v,Vt);return}if(C===Q){i(_,p,v);for(let A=0;A<x.length;A++)ht(x[A],p,v,S);i(u.anchor,p,v);return}if(C===kn){H(u,p,v);return}if(S!==2&&w&1&&I)if(S===0)I.beforeEnter(_),i(_,p,v),xe(()=>I.enter(_),b);else{const{leave:A,delayLeave:F,afterLeave:j}=I,J=()=>{u.ctx.isUnmounted?s(_):i(_,p,v)},ne=()=>{_._isLeaving&&_[lo](!0),A(_,()=>{J(),j&&j()})};F?F(_,J,ne):ne()}else i(_,p,v)},Re=(u,p,v,S=!1,b=!1)=>{const{type:_,props:C,ref:I,children:x,dynamicChildren:w,shapeFlag:M,patchFlag:A,dirs:F,cacheIndex:j,memo:J}=u;if(A===-2&&(b=!1),I!=null&&(tt(),Qt(I,null,v,u,!0),nt()),j!=null&&(p.renderCache[j]=void 0),M&256){p.ctx.deactivate(u);return}const ne=M&1&&F,re=!en(u);let he;if(re&&(he=C&&C.onVnodeBeforeUnmount)&&He(he,p,u),M&6)or(u.component,v,S);else{if(M&128){u.suspense.unmount(v,S);return}ne&&ft(u,null,p,"beforeUnmount"),M&64?u.type.remove(u,p,v,Vt,S):w&&!w.hasOnce&&(_!==Q||A>0&&A&64)?Mt(w,p,v,!1,!0):(_===Q&&A&384||!b&&M&16)&&Mt(x,p,v),S&&Zs(u)}const pe=J!=null&&j==null;(re&&(he=C&&C.onVnodeUnmounted)||ne||pe)&&xe(()=>{he&&He(he,p,u),ne&&ft(u,null,p,"unmounted"),pe&&(u.el=null)},v)},Zs=u=>{const{type:p,el:v,anchor:S,transition:b}=u;if(p===Q){rr(v,S);return}if(p===kn){N(u);return}const _=()=>{s(v),b&&!b.persisted&&b.afterLeave&&b.afterLeave()};if(u.shapeFlag&1&&b&&!b.persisted){const{leave:C,delayLeave:I}=b,x=()=>C(v,_);I?I(u.el,_,x):x()}else _()},rr=(u,p)=>{let v;for(;u!==p;)v=m(u),s(u),u=v;s(p)},or=(u,p,v)=>{const{bum:S,scope:b,job:_,subTree:C,um:I,m:x,a:w}=u;yi(x),yi(w),S&&os(S),b.stop(),_&&(_.flags|=8,Re(C,u,p,v)),I&&xe(I,p),xe(()=>{u.isUnmounted=!0},p)},Mt=(u,p,v,S=!1,b=!1,_=0)=>{for(let C=_;C<u.length;C++)Re(u[C],p,v,S,b)},_n=u=>{if(u.shapeFlag&6)return _n(u.component.subTree);if(u.shapeFlag&128)return u.suspense.next();const p=m(u.anchor||u.el),v=p&&p[ro];return v?m(v):p};let as=!1;const Qs=(u,p,v)=>{let S;u==null?p._vnode&&(Re(p._vnode,null,null,!0),S=p._vnode.component):D(p._vnode||null,u,p,null,null,null,v),p._vnode=u,as||(as=!0,oi(S),ua(),as=!1)},Vt={p:D,um:Re,m:ht,r:Zs,mt:Rt,mc:G,pc:ae,pbc:le,n:_n,o:e};return{render:Qs,hydrate:void 0,createApp:Lo(Qs)}}function fs({type:e,props:t},n){return n==="svg"&&e==="foreignObject"||n==="mathml"&&e==="annotation-xml"&&t&&t.encoding&&t.encoding.includes("html")?void 0:n}function pt({effect:e,job:t},n){n?(e.flags|=32,t.flags|=4):(e.flags&=-33,t.flags&=-5)}function Ko(e,t){return(!e||e&&!e.pendingBranch)&&t&&!t.persisted}function Na(e,t,n=!1){const i=e.children,s=t.children;if(O(i)&&O(s))for(let a=0;a<i.length;a++){const r=i[a];let o=s[a];o.shapeFlag&1&&!o.dynamicChildren&&((o.patchFlag<=0||o.patchFlag===32)&&(o=s[a]=Je(s[a]),o.el=r.el),!n&&o.patchFlag!==-2&&Na(r,o)),o.type===Zn&&(o.patchFlag===-1&&(o=s[a]=Je(o)),o.el=r.el),o.type===ct&&!o.el&&(o.el=r.el)}}function qo(e){const t=e.slice(),n=[0];let i,s,a,r,o;const l=e.length;for(i=0;i<l;i++){const d=e[i];if(d!==0){if(s=n[n.length-1],e[s]<d){t[i]=s,n.push(i);continue}for(a=0,r=n.length-1;a<r;)o=a+r>>1,e[n[o]]<d?a=o+1:r=o;d<e[n[a]]&&(a>0&&(t[i]=n[a-1]),n[a]=i)}}for(a=n.length,r=n[a-1];a-- >0;)n[a]=r,r=t[r];return n}function Ba(e){const t=e.subTree.component;if(t)return t.asyncDep&&!t.asyncResolved?t:Ba(t)}function yi(e){if(e)for(let t=0;t<e.length;t++)e[t].flags|=8}function Fa(e){if(e.placeholder)return e.placeholder;const t=e.component;return t?Fa(t.subTree):null}const Ra=e=>e.__isSuspense;function Yo(e,t){t&&t.pendingBranch?O(e)?t.effects.push(...e):t.effects.push(e):Qr(e)}const Q=Symbol.for("v-fgt"),Zn=Symbol.for("v-txt"),ct=Symbol.for("v-cmt"),kn=Symbol.for("v-stc"),nn=[];let ke=null;function k(e=!1){nn.push(ke=e?null:[])}function Xo(){nn.pop(),ke=nn[nn.length-1]||null}let un=1;function Fn(e,t=!1){un+=e,e<0&&ke&&t&&(ke.hasOnce=!0)}function Ma(e){return e.dynamicChildren=un>0?ke||Tt:null,Xo(),un>0&&ke&&ke.push(e),e}function T(e,t,n,i,s,a){return Ma(f(e,t,n,i,s,a,!0))}function qs(e,t,n,i,s){return Ma(q(e,t,n,i,s,!0))}function Rn(e){return e?e.__v_isVNode===!0:!1}function Ot(e,t){return e.type===t.type&&e.key===t.key}const Va=({key:e})=>e??null,An=({ref:e,ref_key:t,ref_for:n})=>(typeof e=="number"&&(e=""+e),e!=null?ce(e)||de(e)||$(e)?{i:Ce,r:e,k:t,f:!!n}:e:null);function f(e,t=null,n=null,i=0,s=null,a=e===Q?0:1,r=!1,o=!1){const l={__v_isVNode:!0,__v_skip:!0,type:e,props:t,key:t&&Va(t),ref:t&&An(t),scopeId:fa,slotScopeIds:null,children:n,component:null,suspense:null,ssContent:null,ssFallback:null,dirs:null,transition:null,el:null,anchor:null,target:null,targetStart:null,targetAnchor:null,staticCount:0,shapeFlag:a,patchFlag:i,dynamicProps:s,dynamicChildren:null,appContext:null,ctx:Ce};return o?(Ys(l,n),a&128&&e.normalize(l)):n&&(l.shapeFlag|=ce(n)?8:16),un>0&&!r&&ke&&(l.patchFlag>0||a&6)&&l.patchFlag!==32&&ke.push(l),l}const q=Jo;function Jo(e,t=null,n=null,i=0,s=null,a=!1){if((!e||e===So)&&(e=ct),Rn(e)){const o=Ft(e,t,!0);return n&&Ys(o,n),un>0&&!a&&ke&&(o.shapeFlag&6?ke[ke.indexOf(e)]=o:ke.push(o)),o.patchFlag=-2,o}if(cl(e)&&(e=e.__vccOpts),t){t=Zo(t);let{class:o,style:l}=t;o&&!ce(o)&&(t.class=R(o)),te(l)&&(qn(l)&&!O(l)&&(l=ye({},l)),t.style=Te(l))}const r=ce(e)?1:Ra(e)?128:oo(e)?64:te(e)?4:$(e)?2:0;return f(e,t,n,i,s,r,a,!0)}function Zo(e){return e?qn(e)||ka(e)?ye({},e):e:null}function Ft(e,t,n=!1,i=!1){const{props:s,ref:a,patchFlag:r,children:o,transition:l}=e,d=t?el(s||{},t):s,c={__v_isVNode:!0,__v_skip:!0,type:e.type,props:d,key:d&&Va(d),ref:t&&t.ref?n&&a?O(a)?a.concat(An(t)):[a,An(t)]:An(t):a,scopeId:e.scopeId,slotScopeIds:e.slotScopeIds,children:o,target:e.target,targetStart:e.targetStart,targetAnchor:e.targetAnchor,staticCount:e.staticCount,shapeFlag:e.shapeFlag,patchFlag:t&&e.type!==Q?r===-1?16:r|16:r,dynamicProps:e.dynamicProps,dynamicChildren:e.dynamicChildren,appContext:e.appContext,dirs:e.dirs,transition:l,component:e.component,suspense:e.suspense,ssContent:e.ssContent&&Ft(e.ssContent),ssFallback:e.ssFallback&&Ft(e.ssFallback),placeholder:e.placeholder,el:e.el,anchor:e.anchor,ctx:e.ctx,ce:e.ce};return l&&i&&$s(c,l.clone(c)),c}function sn(e=" ",t=0){return q(Zn,null,e,t)}function Qo(e,t){const n=q(kn,null,e);return n.staticCount=t,n}function dt(e="",t=!1){return t?(k(),qs(ct,null,e)):q(ct,null,e)}function $e(e){return e==null||typeof e=="boolean"?q(ct):O(e)?q(Q,null,e.slice()):Rn(e)?Je(e):q(Zn,null,String(e))}function Je(e){return e.el===null&&e.patchFlag!==-1||e.memo?e:Ft(e)}function Ys(e,t){let n=0;const{shapeFlag:i}=e;if(t==null)t=null;else if(O(t))n=16;else if(typeof t=="object")if(i&65){const s=t.default;s&&(s._c&&(s._d=!1),Ys(e,s()),s._c&&(s._d=!0));return}else{n=32;const s=t._;!s&&!ka(t)?t._ctx=Ce:s===3&&Ce&&(Ce.slots._===1?t._=1:(t._=2,e.patchFlag|=1024))}else $(t)?(t={default:t,_ctx:Ce},n=32):(t=String(t),i&64?(n=16,t=[sn(t)]):n=8);e.children=t,e.shapeFlag|=n}function el(...e){const t={};for(let n=0;n<e.length;n++){const i=e[n];for(const s in i)if(s==="class")t.class!==i.class&&(t.class=R([t.class,i.class]));else if(s==="style")t.style=Te([t.style,i.style]);else if(Hn(s)){const a=t[s],r=i[s];r&&a!==r&&!(O(a)&&a.includes(r))?t[s]=a?[].concat(a,r):r:r==null&&a==null&&!On(s)&&(t[s]=r)}else s!==""&&(t[s]=i[s])}return t}function He(e,t,n,i=null){Ke(e,t,7,[n,i])}const tl=Sa();let nl=0;function sl(e,t,n){const i=e.type,s=(t?t.appContext:e.appContext)||tl,a={uid:nl++,vnode:e,type:i,parent:t,appContext:s,root:null,next:null,subTree:null,effect:null,update:null,job:null,scope:new Wi(!0),render:null,proxy:null,exposed:null,exposeProxy:null,withProxy:null,provides:t?t.provides:Object.create(s.provides),ids:t?t.ids:["",0,0],accessCache:null,renderCache:[],components:null,directives:null,propsOptions:Da(i,s),emitsOptions:Ea(i,s),emit:null,emitted:null,propsDefaults:se,inheritAttrs:i.inheritAttrs,ctx:se,data:se,props:se,attrs:se,slots:se,refs:se,setupState:se,setupContext:null,suspense:n,suspenseId:n?n.pendingId:0,asyncDep:null,asyncResolved:!1,isMounted:!1,isUnmounted:!1,isDeactivated:!1,bc:null,c:null,bm:null,m:null,bu:null,u:null,um:null,bum:null,da:null,a:null,rtg:null,rtc:null,ec:null,sp:null};return a.ctx={_:a},a.root=t?t.root:a,a.emit=No.bind(null,a),e.ce&&e.ce(a),a}let me=null;const Ga=()=>me||Ce;let Mn,As;{const e=Wn(),t=(n,i)=>{let s;return(s=e[n])||(s=e[n]=[]),s.push(i),a=>{s.length>1?s.forEach(r=>r(a)):s[0](a)}};Mn=t("__VUE_INSTANCE_SETTERS__",n=>me=n),As=t("__VUE_SSR_SETTERS__",n=>hn=n)}const mn=e=>{const t=me;return Mn(e),e.scope.on(),()=>{e.scope.off(),Mn(t)}},bi=()=>{me&&me.scope.off(),Mn(null)};function Ha(e){return e.vnode.shapeFlag&4}let hn=!1;function il(e,t=!1,n=!1){t&&As(t);const{props:i,children:s}=e.vnode,a=Ha(e);Go(e,i,a,t),$o(e,s,n||t);const r=a?al(e,t):void 0;return t&&As(!1),r}function al(e,t){const n=e.type;e.accessCache=Object.create(null),e.proxy=new Proxy(e.ctx,xo);const{setup:i}=n;if(i){tt();const s=e.setupContext=i.length>1?ol(e):null,a=mn(e),r=vn(i,e,0,[e.props,s]),o=Vi(r);if(nt(),a(),(o||e.sp)&&!en(e)&&va(e),o){if(r.then(bi,bi),t)return r.then(l=>{_i(e,l)}).catch(l=>{Yn(l,e,0)});e.asyncDep=r}else _i(e,r)}else Oa(e)}function _i(e,t,n){$(t)?e.type.__ssrInlineRender?e.ssrRender=t:e.render=t:te(t)&&(e.setupState=la(t)),Oa(e)}function Oa(e,t,n){const i=e.type;e.render||(e.render=i.render||Ue);{const s=mn(e);tt();try{Io(e)}finally{nt(),s()}}}const rl={get(e,t){return ve(e,"get",""),e[t]}};function ol(e){const t=n=>{e.exposed=n||{}};return{attrs:new Proxy(e.attrs,rl),slots:e.slots,emit:e.emit,expose:t}}function Qn(e){return e.exposed?e.exposeProxy||(e.exposeProxy=new Proxy(la(yt(e.exposed)),{get(t,n){if(n in t)return t[n];if(n in tn)return tn[n](e)},has(t,n){return n in t||n in tn}})):e.proxy}function ll(e,t=!0){return $(e)?e.displayName||e.name:e.name||t&&e.__name}function cl(e){return $(e)&&"__vccOpts"in e}const Ne=(e,t)=>qr(e,t,hn);function dl(e,t,n){try{Fn(-1);const i=arguments.length;return i===2?te(t)&&!O(t)?Rn(t)?q(e,null,[t]):q(e,t):q(e,null,t):(i>3?n=Array.prototype.slice.call(arguments,2):i===3&&Rn(n)&&(n=[n]),q(e,t,n))}finally{Fn(1)}}const ul="3.5.33";/**
* @vue/runtime-dom v3.5.33
* (c) 2018-present Yuxi (Evan) You and Vue contributors
* @license MIT
**/let Ds;const wi=typeof window<"u"&&window.trustedTypes;if(wi)try{Ds=wi.createPolicy("vue",{createHTML:e=>e})}catch{}const ja=Ds?e=>Ds.createHTML(e):e=>e,hl="http://www.w3.org/2000/svg",fl="http://www.w3.org/1998/Math/MathML",Xe=typeof document<"u"?document:null,Si=Xe&&Xe.createElement("template"),pl={insert:(e,t,n)=>{t.insertBefore(e,n||null)},remove:e=>{const t=e.parentNode;t&&t.removeChild(e)},createElement:(e,t,n,i)=>{const s=t==="svg"?Xe.createElementNS(hl,e):t==="mathml"?Xe.createElementNS(fl,e):n?Xe.createElement(e,{is:n}):Xe.createElement(e);return e==="select"&&i&&i.multiple!=null&&s.setAttribute("multiple",i.multiple),s},createText:e=>Xe.createTextNode(e),createComment:e=>Xe.createComment(e),setText:(e,t)=>{e.nodeValue=t},setElementText:(e,t)=>{e.textContent=t},parentNode:e=>e.parentNode,nextSibling:e=>e.nextSibling,querySelector:e=>Xe.querySelector(e),setScopeId(e,t){e.setAttribute(t,"")},insertStaticContent(e,t,n,i,s,a){const r=n?n.previousSibling:t.lastChild;if(s&&(s===a||s.nextSibling))for(;t.insertBefore(s.cloneNode(!0),n),!(s===a||!(s=s.nextSibling)););else{Si.innerHTML=ja(i==="svg"?`<svg>${e}</svg>`:i==="mathml"?`<math>${e}</math>`:e);const o=Si.content;if(i==="svg"||i==="mathml"){const l=o.firstChild;for(;l.firstChild;)o.appendChild(l.firstChild);o.removeChild(l)}t.insertBefore(o,n)}return[r?r.nextSibling:t.firstChild,n?n.previousSibling:t.lastChild]}},gl=Symbol("_vtc");function vl(e,t,n){const i=e[gl];i&&(t=(t?[t,...i]:[...i]).join(" ")),t==null?e.removeAttribute("class"):n?e.setAttribute("class",t):e.className=t}const Vn=Symbol("_vod"),$a=Symbol("_vsh"),fn={name:"show",beforeMount(e,{value:t},{transition:n}){e[Vn]=e.style.display==="none"?"":e.style.display,n&&t?n.beforeEnter(e):jt(e,t)},mounted(e,{value:t},{transition:n}){n&&t&&n.enter(e)},updated(e,{value:t,oldValue:n},{transition:i}){!t!=!n&&(i?t?(i.beforeEnter(e),jt(e,!0),i.enter(e)):i.leave(e,()=>{jt(e,!1)}):jt(e,t))},beforeUnmount(e,{value:t}){jt(e,t)}};function jt(e,t){e.style.display=t?e[Vn]:"none",e[$a]=!t}const ml=Symbol(""),yl=/(?:^|;)\s*display\s*:/;function bl(e,t,n){const i=e.style,s=ce(n);let a=!1;if(n&&!s){if(t)if(ce(t))for(const r of t.split(";")){const o=r.slice(0,r.indexOf(":")).trim();n[o]==null&&Kt(i,o,"")}else for(const r in t)n[r]==null&&Kt(i,r,"");for(const r in n){r==="display"&&(a=!0);const o=n[r];o!=null?wl(e,r,!ce(t)&&t?t[r]:void 0,o)||Kt(i,r,o):Kt(i,r,"")}}else if(s){if(t!==n){const r=i[ml];r&&(n+=";"+r),i.cssText=n,a=yl.test(n)}}else t&&e.removeAttribute("style");Vn in e&&(e[Vn]=a?i.display:"",e[$a]&&(i.display="none"))}const Ei=/\s*!important$/;function Kt(e,t,n){if(O(n))n.forEach(i=>Kt(e,t,i));else if(n==null&&(n=""),t.startsWith("--"))e.setProperty(t,n);else{const i=_l(e,t);Ei.test(n)?e.setProperty(St(i),n.replace(Ei,""),"important"):e[i]=n}}const xi=["Webkit","Moz","ms"],ps={};function _l(e,t){const n=ps[t];if(n)return n;let i=we(t);if(i!=="filter"&&i in e)return ps[t]=i;i=zn(i);for(let s=0;s<xi.length;s++){const a=xi[s]+i;if(a in e)return ps[t]=a}return t}function wl(e,t,n,i){return e.tagName==="TEXTAREA"&&(t==="width"||t==="height")&&ce(i)&&n===i}const Ii="http://www.w3.org/1999/xlink";function Ci(e,t,n,i,s,a=yr(t)){i&&t.startsWith("xlink:")?n==null?e.removeAttributeNS(Ii,t.slice(6,t.length)):e.setAttributeNS(Ii,t,n):n==null||a&&!ji(n)?e.removeAttribute(t):e.setAttribute(t,a?"":Le(n)?String(n):n)}function ki(e,t,n,i,s){if(t==="innerHTML"||t==="textContent"){n!=null&&(e[t]=t==="innerHTML"?ja(n):n);return}const a=e.tagName;if(t==="value"&&a!=="PROGRESS"&&!a.includes("-")){const o=a==="OPTION"?e.getAttribute("value")||"":e.value,l=n==null?e.type==="checkbox"?"on":"":String(n);(o!==l||!("_value"in e))&&(e.value=l),n==null&&e.removeAttribute(t),e._value=n;return}let r=!1;if(n===""||n==null){const o=typeof e[t];o==="boolean"?n=ji(n):n==null&&o==="string"?(n="",r=!0):o==="number"&&(n=0,r=!0)}try{e[t]=n}catch{}r&&e.removeAttribute(s||t)}function Sl(e,t,n,i){e.addEventListener(t,n,i)}function El(e,t,n,i){e.removeEventListener(t,n,i)}const Ai=Symbol("_vei");function xl(e,t,n,i,s=null){const a=e[Ai]||(e[Ai]={}),r=a[t];if(i&&r)r.value=i;else{const[o,l]=Il(t);if(i){const d=a[t]=Al(i,s);Sl(e,o,d,l)}else r&&(El(e,o,r,l),a[t]=void 0)}}const Di=/(?:Once|Passive|Capture)$/;function Il(e){let t;if(Di.test(e)){t={};let i;for(;i=e.match(Di);)e=e.slice(0,e.length-i[0].length),t[i[0].toLowerCase()]=!0}return[e[2]===":"?e.slice(3):St(e.slice(2)),t]}let gs=0;const Cl=Promise.resolve(),kl=()=>gs||(Cl.then(()=>gs=0),gs=Date.now());function Al(e,t){const n=i=>{if(!i._vts)i._vts=Date.now();else if(i._vts<=n.attached)return;Ke(Dl(i,n.value),t,5,[i])};return n.value=e,n.attached=kl(),n}function Dl(e,t){if(O(t)){const n=e.stopImmediatePropagation;return e.stopImmediatePropagation=()=>{n.call(e),e._stopped=!0},t.map(i=>s=>!s._stopped&&i&&i(s))}else return t}const Ti=e=>e.charCodeAt(0)===111&&e.charCodeAt(1)===110&&e.charCodeAt(2)>96&&e.charCodeAt(2)<123,Tl=(e,t,n,i,s,a)=>{const r=s==="svg";t==="class"?vl(e,i,r):t==="style"?bl(e,n,i):Hn(t)?On(t)||xl(e,t,n,i,a):(t[0]==="."?(t=t.slice(1),!0):t[0]==="^"?(t=t.slice(1),!1):Ll(e,t,i,r))?(ki(e,t,i),!e.tagName.includes("-")&&(t==="value"||t==="checked"||t==="selected")&&Ci(e,t,i,r,a,t!=="value")):e._isVueCE&&(Pl(e,t)||e._def.__asyncLoader&&(/[A-Z]/.test(t)||!ce(i)))?ki(e,we(t),i,a,t):(t==="true-value"?e._trueValue=i:t==="false-value"&&(e._falseValue=i),Ci(e,t,i,r))};function Ll(e,t,n,i){if(i)return!!(t==="innerHTML"||t==="textContent"||t in e&&Ti(t)&&$(n));if(t==="spellcheck"||t==="draggable"||t==="translate"||t==="autocorrect"||t==="sandbox"&&e.tagName==="IFRAME"||t==="form"||t==="list"&&e.tagName==="INPUT"||t==="type"&&e.tagName==="TEXTAREA")return!1;if(t==="width"||t==="height"){const s=e.tagName;if(s==="IMG"||s==="VIDEO"||s==="CANVAS"||s==="SOURCE")return!1}return Ti(t)&&ce(n)?!1:t in e}function Pl(e,t){const n=e._def.props;if(!n)return!1;const i=we(t);return Array.isArray(n)?n.some(s=>we(s)===i):Object.keys(n).some(s=>we(s)===i)}const Nl=["ctrl","shift","alt","meta"],Bl={stop:e=>e.stopPropagation(),prevent:e=>e.preventDefault(),self:e=>e.target!==e.currentTarget,ctrl:e=>!e.ctrlKey,shift:e=>!e.shiftKey,alt:e=>!e.altKey,meta:e=>!e.metaKey,left:e=>"button"in e&&e.button!==0,middle:e=>"button"in e&&e.button!==1,right:e=>"button"in e&&e.button!==2,exact:(e,t)=>Nl.some(n=>e[`${n}Key`]&&!t.includes(n))},it=(e,t)=>{if(!e)return e;const n=e._withMods||(e._withMods={}),i=t.join(".");return n[i]||(n[i]=(s,...a)=>{for(let r=0;r<t.length;r++){const o=Bl[t[r]];if(o&&o(s,t))return}return e(s,...a)})},Fl=ye({patchProp:Tl},pl);let Li;function za(){return Li||(Li=Wo(Fl))}const Rl=(...e)=>{za().render(...e)},Ml=(...e)=>{const t=za().createApp(...e),{mount:n}=t;return t.mount=i=>{const s=Gl(i);if(!s)return;const a=t._component;!$(a)&&!a.render&&!a.template&&(a.template=s.innerHTML),s.nodeType===1&&(s.textContent="");const r=n(s,!1,Vl(s));return s instanceof Element&&(s.removeAttribute("v-cloak"),s.setAttribute("data-v-app","")),r},t};function Vl(e){if(e instanceof SVGElement)return"svg";if(typeof MathMLElement=="function"&&e instanceof MathMLElement)return"mathml"}function Gl(e){return ce(e)?document.querySelector(e):e}/*!
 * pinia v2.3.1
 * (c) 2025 Eduardo San Martin Morote
 * @license MIT
 */let Wa;const es=e=>Wa=e,Ua=Symbol();function Ts(e){return e&&typeof e=="object"&&Object.prototype.toString.call(e)==="[object Object]"&&typeof e.toJSON!="function"}var an;(function(e){e.direct="direct",e.patchObject="patch object",e.patchFunction="patch function"})(an||(an={}));function Hl(){const e=Ui(!0),t=e.run(()=>Jt({}));let n=[],i=[];const s=yt({install(a){es(s),s._a=a,a.provide(Ua,s),a.config.globalProperties.$pinia=s,i.forEach(r=>n.push(r)),i=[]},use(a){return this._a?n.push(a):i.push(a),this},_p:n,_a:null,_e:e,_s:new Map,state:t});return s}const Ka=()=>{};function Pi(e,t,n,i=Ka){e.push(t);const s=()=>{const a=e.indexOf(t);a>-1&&(e.splice(a,1),i())};return!n&&Ki()&&_r(s),s}function Ct(e,...t){e.slice().forEach(n=>{n(...t)})}const Ol=e=>e(),Ni=Symbol(),vs=Symbol();function Ls(e,t){e instanceof Map&&t instanceof Map?t.forEach((n,i)=>e.set(i,n)):e instanceof Set&&t instanceof Set&&t.forEach(e.add,e);for(const n in t){if(!t.hasOwnProperty(n))continue;const i=t[n],s=e[n];Ts(s)&&Ts(i)&&e.hasOwnProperty(n)&&!de(i)&&!et(i)?e[n]=Ls(s,i):e[n]=i}return e}const jl=Symbol();function $l(e){return!Ts(e)||!e.hasOwnProperty(jl)}const{assign:ot}=Object;function zl(e){return!!(de(e)&&e.effect)}function Wl(e,t,n,i){const{state:s,actions:a,getters:r}=t,o=n.state.value[e];let l;function d(){o||(n.state.value[e]=s?s():{});const c=zr(n.state.value[e]);return ot(c,a,Object.keys(r||{}).reduce((h,m)=>(h[m]=yt(Ne(()=>{es(n);const E=n._s.get(e);return r[m].call(E,E)})),h),{}))}return l=qa(e,d,t,n,i,!0),l}function qa(e,t,n={},i,s,a){let r;const o=ot({actions:{}},n),l={deep:!0};let d,c,h=[],m=[],E;const B=i.state.value[e];!a&&!B&&(i.state.value[e]={});let D;function W(G){let z;d=c=!1,typeof G=="function"?(G(i.state.value[e]),z={type:an.patchFunction,storeId:e,events:E}):(Ls(i.state.value[e],G),z={type:an.patchObject,payload:G,storeId:e,events:E});const le=D=Symbol();Et().then(()=>{D===le&&(d=!0)}),c=!0,Ct(h,z,i.state.value[e])}const U=a?function(){const{state:z}=n,le=z?z():{};this.$patch(De=>{ot(De,le)})}:Ka;function V(){r.stop(),h=[],m=[],i._s.delete(e)}const H=(G,z="")=>{if(Ni in G)return G[vs]=z,G;const le=function(){es(i);const De=Array.from(arguments),Fe=[],rt=[];function Rt(K){Fe.push(K)}function yn(K){rt.push(K)}Ct(m,{args:De,name:le[vs],store:Y,after:Rt,onError:yn});let ue;try{ue=G.apply(this&&this.$id===e?this:Y,De)}catch(K){throw Ct(rt,K),K}return ue instanceof Promise?ue.then(K=>(Ct(Fe,K),K)).catch(K=>(Ct(rt,K),Promise.reject(K))):(Ct(Fe,ue),ue)};return le[Ni]=!0,le[vs]=z,le},N={_p:i,$id:e,$onAction:Pi.bind(null,m),$patch:W,$reset:U,$subscribe(G,z={}){const le=Pi(h,G,z.detached,()=>De()),De=r.run(()=>Nt(()=>i.state.value[e],Fe=>{(z.flush==="sync"?c:d)&&G({storeId:e,type:an.direct,events:E},Fe)},ot({},l,z)));return le},$dispose:V},Y=Kn(N);i._s.set(e,Y);const P=(i._a&&i._a.runWithContext||Ol)(()=>i._e.run(()=>(r=Ui()).run(()=>t({action:H}))));for(const G in P){const z=P[G];if(de(z)&&!zl(z)||et(z))a||(B&&$l(z)&&(de(z)?z.value=B[G]:Ls(z,B[G])),i.state.value[e][G]=z);else if(typeof z=="function"){const le=H(z,G);P[G]=le,o.actions[G]=z}}return ot(Y,P),ot(Z(Y),P),Object.defineProperty(Y,"$state",{get:()=>i.state.value[e],set:G=>{W(z=>{ot(z,G)})}}),i._p.forEach(G=>{ot(Y,r.run(()=>G({store:Y,app:i._a,pinia:i,options:o})))}),B&&a&&n.hydrate&&n.hydrate(Y.$state,B),d=!0,c=!0,Y}/*! #__NO_SIDE_EFFECTS__ */function Ee(e,t,n){let i,s;const a=typeof t=="function";typeof e=="string"?(i=e,s=a?n:t):(s=e,i=e.id);function r(o,l){const d=no();return o=o||(d?Zt(Ua,null):null),o&&es(o),o=Wa,o._s.has(i)||(a?qa(i,t,s,o):Wl(i,s,o)),o._s.get(i)}return r.$id=i,r}const Ya=Ee("recentFiles",{state:()=>({items:[],loader:null,flashingId:null}),actions:{setItems(e=[],t=null){this.items=Array.isArray(e)?e:[],this.loader=typeof t=="function"?t:null,this.flashingId=null},flash(e){this.flashingId=e==null?null:String(e)},clearFlash(){this.flashingId=null},load(e){e&&this.loader&&this.loader(e.id)}}}),xt=Ee("dropdowns",{state:()=>({openId:"",flashingItemId:""}),getters:{isOpen:e=>t=>e.openId===t},actions:{toggle(e){return this.openId=this.openId===e?"":e,this.openId===e},closeAll(){this.openId=""},flashItem(e,t=50){e&&(this.flashingItemId=String(e),window.setTimeout(()=>{this.flashingItemId===String(e)&&(this.flashingItemId="")},t))}}}),Ul={id:"file-recent-list"},Kl={key:0,href:"#",class:"recent-files-empty"},ql=["title","onClick"],Yl={__name:"RecentFilesMenu",setup(e){const t=Ya(),n=xt();function i(s){t.flash(s.id),window.setTimeout(()=>{t.clearFlash(),n.closeAll(),t.load(s)},50)}return(s,a)=>(k(),T("div",Ul,[g(t).items.length===0?(k(),T("a",Kl,"(no recent files)")):(k(!0),T(Q,{key:1},Se(g(t).items,r=>(k(),T("a",{key:r.id,href:"#",title:r.title||"",class:R({"menu-flash":g(t).flashingId===String(r.id)}),onClick:it(o=>i(r),["stop","prevent"])},ie(r.label),11,ql))),128))]))}};function ge(e,t){e.preventDefault(),e.stopPropagation(),typeof e.stopImmediatePropagation=="function"&&e.stopImmediatePropagation(),typeof t=="function"&&t()}function ts(){return globalThis.window&&globalThis.window.ivyApp}function L(e,...t){const n=ts();if(n&&typeof n[e]=="function")return n[e](...t)}function pn(e){const t=ts();return!!(t&&typeof t[e]=="function")}function Xl(e,t=""){const n=ts();return n&&n.controls&&typeof n.controls.setStatus=="function"?(n.controls.setStatus(e,t),!0):!1}function X(e,t){ge(e,()=>{const n=ts();n&&typeof n.flashAndClose=="function"?n.flashAndClose(e.currentTarget,t):typeof t=="function"&&t()})}function Gn(e,t,n,i){ge(e,()=>{t.toggle(n)&&typeof i=="function"&&i()})}const ms="/static/tutorial/kenmcmil.github.io/ivy/language.html";function Jl(e){const t=String(e||"").trim();return t?/^https?:\/\//i.test(t)||t.startsWith("/")?t:`https://${t}`:""}function Zl(e){const t=String(e||"").trim();if(!t||t==="about:blank")return"";try{const n=new URL(t);return typeof window<"u"&&n.origin===window.location.origin?`${n.pathname}${n.search}${n.hash}`:t}catch{return t}}const qe=Ee("layout",{state:()=>({tutorialVisible:!0,tutorialUrl:ms,tutorialInput:ms,tutorialHistory:[ms],tutorialHistoryIndex:0,tutorialFrameKey:0,tutorialButtonFlashing:!1,argPanelWidth:null,statePanelWidth:null,detailsHeight:null,editorWidth:null,tutorialHeight:null}),getters:{canGoBack:e=>e.tutorialHistoryIndex>0,canGoForward:e=>e.tutorialHistoryIndex<e.tutorialHistory.length-1,argPanelStyle:e=>e.argPanelWidth?{flex:`0 0 ${e.argPanelWidth}px`}:{},statePanelStyle:e=>e.statePanelWidth?{flex:`0 0 ${e.statePanelWidth}px`}:{},detailsPanelStyle:e=>e.detailsHeight?{flex:`0 0 ${e.detailsHeight}px`,height:`${e.detailsHeight}px`}:{},editorPanelStyle:e=>e.editorWidth?{flex:`0 0 ${e.editorWidth}px`}:{},tutorialPanelStyle:e=>({display:e.tutorialVisible?"":"none",...e.tutorialHeight?{flex:`0 0 ${e.tutorialHeight}px`}:{}})},actions:{setTutorialVisible(e){this.tutorialVisible=!!e},flashTutorialButton(e=1200){this.tutorialButtonFlashing=!0,window.setTimeout(()=>{this.tutorialButtonFlashing=!1},Number(e)||1200)},setTutorialInput(e){this.tutorialInput=String(e||"")},navigateTutorial(e){const t=Jl(e);t&&(this.tutorialHistoryIndex<this.tutorialHistory.length-1&&(this.tutorialHistory=this.tutorialHistory.slice(0,this.tutorialHistoryIndex+1)),this.tutorialHistory[this.tutorialHistoryIndex]!==t&&(this.tutorialHistory.push(t),this.tutorialHistoryIndex=this.tutorialHistory.length-1),this.tutorialInput=t,this.tutorialUrl=t)},goTutorialBack(){if(!this.canGoBack)return;this.tutorialHistoryIndex-=1;const e=this.tutorialHistory[this.tutorialHistoryIndex];this.tutorialInput=e,this.tutorialUrl=e,this.tutorialFrameKey+=1},goTutorialForward(){if(!this.canGoForward)return;this.tutorialHistoryIndex+=1;const e=this.tutorialHistory[this.tutorialHistoryIndex];this.tutorialInput=e,this.tutorialUrl=e,this.tutorialFrameKey+=1},reloadTutorial(){const e=this.tutorialHistory[this.tutorialHistoryIndex]||this.tutorialUrl;e&&(this.tutorialInput=e,this.tutorialUrl=e,this.tutorialFrameKey+=1)},recordTutorialLoad(e){const t=Zl(e);t&&(this.tutorialHistory[this.tutorialHistoryIndex]!==t&&(this.tutorialHistoryIndex<this.tutorialHistory.length-1&&(this.tutorialHistory=this.tutorialHistory.slice(0,this.tutorialHistoryIndex+1)),this.tutorialHistory.push(t),this.tutorialHistoryIndex=this.tutorialHistory.length-1),this.tutorialInput=t,this.tutorialUrl=t)},setArgPanelWidth(e){this.argPanelWidth=Math.max(150,Number(e)||150)},setStatePanelWidth(e){this.statePanelWidth=Math.max(200,Number(e)||200)},setDetailsHeight(e){this.detailsHeight=Math.max(72,Number(e)||72)},setEditorWidth(e){this.editorWidth=Math.max(200,Number(e)||200)},setTutorialHeight(e){this.tutorialHeight=Math.max(80,Number(e)||80)}}});class Xa{async createSession(){throw new Error("createSession is not implemented")}async loadModel(t){throw new Error("loadModel is not implemented")}async getArg(t={}){throw new Error("getArg is not implemented")}async getConcept(t={}){throw new Error("getConcept is not implemented")}async getMenus(){throw new Error("getMenus is not implemented")}async check(t={}){throw new Error("check is not implemented")}async runAction(t){throw new Error("runAction is not implemented")}async runArgAction(t){throw new Error("runArgAction is not implemented")}async getToggles(){throw new Error("getToggles is not implemented")}async setToggles(t){throw new Error("setToggles is not implemented")}async requestSession(t,n={}){throw new Error("requestSession is not implemented")}async fetchSession(t,n={}){throw new Error("fetchSession is not implemented")}subscribeEvents(t,n){return()=>{}}}class Ja{constructor({baseURL:t="",fetchImpl:n=typeof globalThis.fetch=="function"?globalThis.fetch.bind(globalThis):void 0}={}){if(typeof n!="function")throw new Error("IvyHttpClient requires a fetch implementation");this.baseURL=t,this.fetchImpl=n}async request(t,n={}){var a,r;const i=await this.fetch(t,n);return(((r=(a=i.headers)==null?void 0:a.get)==null?void 0:r.call(a,"content-type"))||"").includes("application/json")?i.json():i.text()}async fetch(t,n={}){const i=await this.fetchImpl(this.baseURL+t,n);if(!i.ok){let s="";try{s=await i.text()}catch{s=""}throw new Error(`API error ${i.status}: ${s||i.statusText}`)}return i}}function Ql(e={}){const t=new FormData;if(e.file)return t.append("file",e.file),t;const n=e.content||"",i=e.filename||e.path||"model.ivy",s=new Blob([n],{type:"text/plain"});return t.append("file",s,i),t}class Ps extends Xa{constructor({client:t=new Ja,eventSourceFactory:n=i=>new EventSource(i)}={}){super(),this.client=t,this.eventSourceFactory=n,this.sessionId=null,this.eventSource=null}sessionPath(t){if(!this.sessionId)throw new Error("No Ivy session has been created yet");return`/api/session/${this.sessionId}${t}`}async requestSession(t,n={}){return this.client.request(this.sessionPath(t),n)}async fetchSession(t,n={}){return this.client.fetch(this.sessionPath(t),n)}async createSession(){const t=await this.client.request("/api/session/new",{method:"POST"});return this.sessionId=t.session_id,this.sessionId}async loadModel(t={}){return this.client.request(this.sessionPath("/load"),{method:"POST",body:Ql(t)})}async getArg({sheetId:t}={}){const n=t?`/arg?sheet=${encodeURIComponent(t)}`:"/arg";return this.client.request(this.sessionPath(n))}async getConcept({sheetId:t,nodeId:n,stateId:i}={}){const s=new URLSearchParams;t&&s.set("sheet",t);const a=n??i;a!=null&&s.set("node",String(a));const r=s.toString();return this.client.request(this.sessionPath(`/concept${r?`?${r}`:""}`))}async getMenus(){return this.client.request(this.sessionPath("/menus"))}async check({mode:t,...n}={}){const i={...n};return t&&(i.mode=t),this.client.request(this.sessionPath("/check"),{method:"POST",headers:{"Content-Type":"application/json"},body:JSON.stringify(i)})}async runAction(t={}){return this.client.request(this.sessionPath("/action"),{method:"POST",headers:{"Content-Type":"application/json"},body:JSON.stringify(t)})}async runArgAction(t={}){return this.client.request(this.sessionPath("/arg/action"),{method:"POST",headers:{"Content-Type":"application/json"},body:JSON.stringify(t)})}async getToggles(){return this.client.request(this.sessionPath("/toggles"))}async setToggles(t={}){return this.client.request(this.sessionPath("/toggles"),{method:"POST",headers:{"Content-Type":"application/json"},body:JSON.stringify(t)})}subscribeEvents(t,n){if(!this.sessionId)throw new Error("No Ivy session has been created yet");this.closeEvents();const i=this.eventSourceFactory(this.sessionPath("/events"));return this.eventSource=i,i.onmessage=s=>{if(t)try{t(JSON.parse(s.data))}catch{t(s.data)}},i.onerror=s=>{n&&n(s)},()=>this.closeEvents()}closeEvents(){this.eventSource&&typeof this.eventSource.close=="function"&&this.eventSource.close(),this.eventSource=null}}function gt(e={}){return{method:"POST",headers:{"Content-Type":"application/json"},body:JSON.stringify(e)}}class Za{constructor(t){this.engine=t,this.onConnectionLost=null,this._sseRetries=0,this._sseMaxRetries=5,this._sseClosed=!0,this._sseTimer=null,this._sseOnEvent=null,this._unsubscribe=null}get sessionId(){return this.engine.sessionId}set sessionId(t){this.engine.sessionId=t}get eventSource(){return this.engine.eventSource}set eventSource(t){this.engine.eventSource=t}async createSession(){return this.engine.createSession()}async loadFile(t){return this.engine.loadModel({file:t})}async reloadContent(t,n){return this.engine.loadModel({content:t,filename:n||"model.ivy"})}async getARG(){return this.engine.getArg()}async getMenus(){return this.engine.getMenus()}async getConceptGraph(t,n){return this.engine.getConcept({nodeId:t,sheetId:n})}async getProofGraph(){return this.engine.requestSession("/proof")}async executeAction(t,n){return this.engine.runAction({action:t,args:n||{}})}async splitConcept(t,n){return this.engine.requestSession("/concept/split",gt({concept:t,split_by:n}))}async supposeEmpty(t){return this.engine.requestSession("/concept/empty",gt({concept:t}))}async removeConcept(t){return this.engine.requestSession("/concept/remove",gt({concept:t}))}async undo(){return this.engine.requestSession("/concept/undo",{method:"POST"})}async materializeNode(t){return this.engine.requestSession("/concept/materialize",gt({concept:t,type:"node"}))}async materializeEdge(t,n,i,s){return this.engine.requestSession("/concept/materialize",gt({relation:t,source:n,target:i,type:"edge",positive:s}))}async addProjection(t,n){return this.engine.requestSession("/concept/projection",gt({name:t,concept:n}))}async runCheck(t,n={}){return this.engine.check({mode:t,...n})}async resetDomain(){return this.engine.requestSession("/concept/reset",{method:"POST"})}async diagramDomain(){return this.engine.requestSession("/concept/diagram",{method:"POST"})}async getToggles(){return this.engine.getToggles()}async setToggles(t){return this.engine.setToggles(t)}async argNodeAction(t,n,i){return this.engine.runArgAction({node:t,action:n,args:i||{}})}async proofGoalAction(t,n){return this.engine.requestSession("/proof/action",gt({goal:t,action:n}))}async saveSession(){const t=await this.engine.fetchSession("/save");if(!t.ok)throw new Error(`Save failed: ${t.statusText||t.status||"unknown error"}`);return t.blob()}connectEvents(t){if(this.disconnectEvents(),typeof this.engine.eventSourceFactory!="function"||typeof this.engine.sessionPath!="function"){this._unsubscribe=this.engine.subscribeEvents(t,()=>{this.onConnectionLost&&this.onConnectionLost()});return}this._sseRetries=0,this._sseClosed=!1,this._sseOnEvent=t,this._sseConnect()}_sseConnect(){if(this._sseClosed||!this.sessionId)return;const t=this.engine.eventSourceFactory(this.engine.sessionPath("/events"));this.eventSource=t,t.onopen=()=>{this._sseRetries=0},t.onmessage=n=>{try{this._sseOnEvent&&this._sseOnEvent(JSON.parse(n.data))}catch(i){console.error("Failed to parse SSE event:",i,n.data)}},t.onerror=()=>{if(t.close(),this._sseClosed)return;if(this._sseRetries+=1,this._sseRetries>this._sseMaxRetries){console.error("SSE: max retries exceeded, giving up"),this.onConnectionLost&&this.onConnectionLost();return}const n=Math.min(1e3*Math.pow(2,this._sseRetries-1),16e3);console.warn(`SSE: reconnect attempt ${this._sseRetries}/${this._sseMaxRetries} in ${n}ms`),this._sseTimer=window.setTimeout(()=>{this._sseConnect()},n)}}disconnectEvents(){this._sseClosed=!0,this._unsubscribe&&(this._unsubscribe(),this._unsubscribe=null),this._sseTimer&&(window.clearTimeout(this._sseTimer),this._sseTimer=null),this.eventSource&&typeof this.eventSource.close=="function"&&this.eventSource.close(),this.eventSource=null}}class ec extends Xa{constructor(){super(),this.kind="wanix"}unavailable(){throw new Error("Wanix Ivy engine is not implemented yet")}async createSession(){this.unavailable()}async loadModel(){this.unavailable()}async getArg(){this.unavailable()}async getConcept(){this.unavailable()}async getMenus(){this.unavailable()}async check(){this.unavailable()}async runAction(){this.unavailable()}async runArgAction(){this.unavailable()}async getToggles(){this.unavailable()}async setToggles(){this.unavailable()}async requestSession(){this.unavailable()}async fetchSession(){this.unavailable()}}const wt=Ee("engine",{state:()=>({kind:"hosted-go",engine:yt(new Ps)}),actions:{useHostedGo(e={}){return this.kind="hosted-go",this.engine=yt(new Ps(e)),this.engine},useWanix(){return this.kind="wanix",this.engine=yt(new ec),this.engine},setEngine(e,t){return this.kind=e,this.engine=yt(t),this.engine}}}),tc=["induction","pdr","concrete","abstract","bounded"],ns=Ee("session",{state:()=>({sessionId:"",sessionDisplay:"",mode:"pdr",loadedFileDisplay:"",loadedFileTitle:"",status:"Ready",statusLevel:"",loading:!1,loadingMessage:"Loading...",saveAsNoticeVisible:!1,events:[],unsubscribeEvents:null}),actions:{setStatus(e,t=""){this.status=e,this.statusLevel=t},setSessionId(e=""){this.sessionId=String(e||""),this.sessionDisplay=this.sessionId?`Session: ${this.sessionId}`:""},setMode(e){const t=String(e||"");this.mode=tc.includes(t)?t:"pdr"},setLoadedFile(e="",t=""){const n=t||e||"";this.loadedFileDisplay=n,this.loadedFileTitle=n},showLoading(e="Loading..."){this.loading=!0,this.loadingMessage=e||"Loading..."},hideLoading(){this.loading=!1},setSaveAsNoticeVisible(e){this.saveAsNoticeVisible=!!e},async createSession(){const e=wt().engine;return this.setStatus("Initializing..."),this.setSessionId(await e.createSession()),this.setStatus("Ready"),this.sessionId},subscribeEvents(){const e=wt().engine;return this.unsubscribeEvents&&(this.unsubscribeEvents(),this.unsubscribeEvents=null),this.unsubscribeEvents=e.subscribeEvents(t=>this.events.push(t),()=>this.setStatus("Server connection lost","error")),this.unsubscribeEvents},closeEvents(){this.unsubscribeEvents&&this.unsubscribeEvents(),this.unsubscribeEvents=null}}}),nc={id:"menubar"},sc={class:"menu-group"},ic={id:"file-menu",class:"dropdown-content"},ac={class:"menu-group"},rc=["value"],oc={class:"menu-group"},lc=["title"],cc={class:"menu-group menu-right"},dc={__name:"Menubar",setup(e){const t=qe(),n=xt(),i=ns();function s(a){return{"menu-flash":n.flashingItemId===a}}return(a,r)=>(k(),T("div",nc,[f("div",sc,[f("div",{class:R(["dropdown",{open:g(n).isOpen("file-menu")}])},[f("span",{class:"panel-menu","data-dropdown":"file-menu",onClick:r[0]||(r[0]=o=>g(Gn)(o,g(n),"file-menu",()=>g(L)("populateRecentFiles")))},"File"),f("div",ic,[f("a",{href:"#",id:"file-load",class:R(s("file-load")),onClick:r[1]||(r[1]=o=>g(X)(o,()=>g(L)("chooseAndLoadModelFile")))},"Load...",2),f("a",{href:"#",id:"file-open-event-trace",class:R(s("file-open-event-trace")),onClick:r[2]||(r[2]=o=>g(X)(o,()=>g(L)("chooseAndLoadEventTraceFile")))},"Open Event Trace...",2),f("a",{href:"#",id:"file-save-as",class:R(s("file-save-as")),onClick:r[3]||(r[3]=o=>g(X)(o,()=>g(L)("saveAs")))},"Save as...",2),f("a",{href:"#",id:"file-download",class:R(s("file-download")),onClick:r[4]||(r[4]=o=>g(X)(o,()=>g(L)("downloadModel")))},"Download current model",2),f("a",{href:"#",id:"file-save-analysis-state",class:R(s("file-save-analysis-state")),onClick:r[5]||(r[5]=o=>g(X)(o,()=>g(L)("saveAnalysisState")))},"Save Analysis State...",2),f("a",{href:"#",id:"file-load-analysis-state",class:R(s("file-load-analysis-state")),onClick:r[6]||(r[6]=o=>g(X)(o,()=>g(L)("chooseAndLoadAnalysisStateFile")))},"Load Analysis State...",2),f("a",{href:"#",id:"file-save-invariant",class:R(s("file-save-invariant")),onClick:r[7]||(r[7]=o=>g(X)(o,()=>g(L)("saveInvariant")))},"Save Invariant...",2),r[16]||(r[16]=f("div",{class:"dropdown-sep"},null,-1)),q(Yl),r[17]||(r[17]=f("div",{class:"dropdown-sep"},null,-1)),r[18]||(r[18]=f("div",{class:"menu-spacer"},null,-1)),f("a",{href:"#",id:"file-new",class:R(s("file-new")),onClick:r[8]||(r[8]=o=>g(X)(o,()=>g(L)("newModel")))},"New Model",2)])],2)]),f("div",ac,[r[20]||(r[20]=f("span",{class:"menu-label"},"Mode",-1)),f("select",{id:"mode-select",value:g(i).mode,onChange:r[9]||(r[9]=o=>g(i).setMode(o.target.value))},[...r[19]||(r[19]=[Qo('<option value="induction">Induction</option><option value="pdr">PDR</option><option value="concrete">Concrete</option><option value="abstract">Abstract</option><option value="bounded">Bounded</option>',5)])],40,rc)]),f("div",oc,[f("button",{id:"btn-check",class:"menu-btn action-btn",onClick:r[10]||(r[10]=o=>g(ge)(o,()=>g(L)("runCheck")))},"Check"),f("button",{id:"btn-show-reachable",class:"menu-btn",onClick:r[11]||(r[11]=o=>g(ge)(o,()=>g(L)("showReachableStates")))},"Show Reachable"),f("button",{id:"btn-undo",class:"menu-btn",onClick:r[12]||(r[12]=o=>g(ge)(o,()=>g(L)("doUndo")))},"Undo"),f("button",{id:"btn-reset-domain",class:"menu-btn",onClick:r[13]||(r[13]=o=>g(ge)(o,()=>g(L)("resetDomain")))},"Reset Domain"),f("button",{id:"btn-diagram-domain",class:"menu-btn",onClick:r[14]||(r[14]=o=>g(ge)(o,()=>g(L)("diagramDomain")))},"Diagram Domain"),f("span",{id:"loaded-file",class:"loaded-file",title:g(i).loadedFileTitle},ie(g(i).loadedFileDisplay),9,lc)]),f("div",cc,[f("button",{id:"btn-toggle-tutorial",class:R(["menu-btn",{"btn-flash":g(t).tutorialButtonFlashing}]),onClick:r[15]||(r[15]=o=>g(ge)(o,()=>g(L)("toggleTutorial")))},ie(g(t).tutorialVisible?"Hide Tutorial":"Show Tutorial"),3)])]))}},uc=["sublime","emacs","vim"],Qa=Ee("editor",{state:()=>({path:"",content:"",savedContent:"",saveState:"idle",hasFile:!1,keymap:"sublime",reopenLastVisible:!1,reopenLastLabel:""}),getters:{dirty:e=>e.content!==e.savedContent,displayName:e=>e.path||"(unsaved file)",label(e){const t=e.path||"(unsaved file)";return e.path?e.saveState==="saving"?`${t} [saving...]`:e.content!==e.savedContent?`** ${t}`:`${t} [saved]`:e.saveState==="saving"?`${t} [saving...]`:e.content!==e.savedContent?`** ${t}`:t}},actions:{load({path:e="",content:t=""}={}){this.path=e,this.content=t,this.savedContent=t,this.hasFile=!!e,this.saveState="idle"},edit(e){this.content=e,this.saveState==="saved"&&(this.saveState="idle")},markSaving(){this.saveState="saving"},markSaved(e=this.content){this.savedContent=e,this.saveState="saved"},markSaveFailed(){this.saveState="idle"},setKeymap(e){const t=String(e||"");this.keymap=uc.includes(t)?t:"sublime"},setReopenLastFileButton(e,t=""){this.reopenLastVisible=!!e,this.reopenLastLabel=this.reopenLastVisible?String(t||"Re-open last file"):""},applyLegacySnapshot({path:e="",content:t="",savedContent:n="",saveInProgress:i=!1}={}){this.path=e,this.content=t,this.savedContent=n,this.hasFile=!!e,this.saveState=i?"saving":"idle"}}}),hc={class:"panel-header"},fc={class:"editor-title-row"},pc=["title"],gc={class:"panel-header-actions"},vc={class:"keymap-radios"},mc=["checked"],yc=["checked"],bc=["checked"],_c={class:"editor-text-shell"},wc={key:0,class:"ivy-save-editor-sheen","aria-hidden":"true"},Sc={__name:"EditorPane",setup(e){const t=Qa(),n=qe();function i(s){t.setKeymap(s),L("setEditorKeymap",s)}return(s,a)=>(k(),T("div",{id:"editor-panel",style:Te(g(n).editorPanelStyle)},[f("div",hc,[f("div",fc,[a[5]||(a[5]=f("strong",{class:"column-title"},"Editing:",-1)),f("span",{id:"model-editor-label",class:"editor-path",title:`Editing: ${g(t).label}`},ie(g(t).label),9,pc)]),f("div",gc,[f("button",{id:"file-close-current",class:"editor-close-btn",title:"Close current file",onClick:a[0]||(a[0]=r=>g(ge)(r,()=>g(L)("closeCurrentFile")))},"x"),dn(f("button",{id:"file-reopen-last",class:"reopen-last-btn",onClick:a[1]||(a[1]=r=>g(ge)(r,()=>g(L)("reopenLastFile")))},ie(g(t).reopenLastLabel),513),[[fn,g(t).reopenLastVisible]]),f("span",vc,[f("label",null,[f("input",{type:"radio",name:"keymap",value:"sublime",checked:g(t).keymap==="sublime",onChange:a[2]||(a[2]=r=>i("sublime"))},null,40,mc),a[6]||(a[6]=sn(" Sublime",-1))]),f("label",null,[f("input",{type:"radio",name:"keymap",value:"emacs",checked:g(t).keymap==="emacs",onChange:a[3]||(a[3]=r=>i("emacs"))},null,40,yc),a[7]||(a[7]=sn(" Emacs-ish",-1))]),f("label",null,[f("input",{type:"radio",name:"keymap",value:"vim",checked:g(t).keymap==="vim",onChange:a[4]||(a[4]=r=>i("vim"))},null,40,bc),a[8]||(a[8]=sn(" Vim",-1))]),a[9]||(a[9]=f("a",{class:"keymap-docs-link",href:"https://codemirror.net/5/doc/manual.html#keymaps",target:"_blank",rel:"noopener noreferrer"},"keymap docs",-1))])])]),f("div",_c,[a[10]||(a[10]=f("textarea",{id:"model-editor",class:"model-editor-text",spellcheck:"false",placeholder:"Load an .ivy file to see its source here..."},null,-1)),g(t).saveState==="saving"?(k(),T("div",wc)):dt("",!0)])],4))}};function Bi(e,t){return String(e).toLowerCase().replace(/[^a-z0-9]+/g,"-").replace(/^-|-$/g,"")||String(t)}function Ec(e,t){return e&&e.type==="separator"?{...e,type:"separator",key:`sep-${t}`}:{...e||{},type:e&&e.type||"item",key:`item-${t}-${e&&(e.action||e.label)||""}`,label:e&&(e.label||e.action)||""}}function xc(e,t,n){const i=t&&t.label||"Menu";return{...t||{},key:`${e}-${n}-${Bi(i,"menu")}`,label:i,contentId:`dynamic-${e}-${n}-${Bi(i,"menu")}`,items:Array.isArray(t&&t.items)?t.items.map(Ec):[]}}const Xs=Ee("menuDescriptor",{state:()=>({regions:{arg:[],concept:[]},dispatchers:{},openKey:""}),getters:{menusFor:e=>t=>e.regions[t]||[],isOpen:e=>(t,n)=>e.openKey===`${t}:${n}`},actions:{setRegion(e,t=[],n=null){this.regions[e]=Array.isArray(t)?t.map((i,s)=>xc(e,i,s)):[],typeof n=="function"?this.dispatchers[e]=n:delete this.dispatchers[e],this.openKey=""},toggle(e,t){const n=`${e}:${t}`;this.openKey=this.openKey===n?"":n},closeAll(){this.openKey=""},runItem(e,t){if(!t||t.enabled===!1)return Promise.resolve({ok:!1,error:"disabled action"});this.closeAll();const n=this.dispatchers[e];return typeof n=="function"?n(t):Promise.resolve({ok:!1,error:"no dispatcher"})}}}),Ic=["data-dynamic-menu-region"],Cc=["data-dropdown","onClick"],kc=["id"],Ac={key:0,class:"dropdown-sep"},Dc=["aria-disabled","data-menu-action","data-menu-dispatch","onClick"],er={__name:"DynamicMenuRegion",props:{region:{type:String,required:!0}},setup(e){const t=e,n=Xs(),i=xt(),s=Ne(()=>n.menusFor(t.region));function a(){i.closeAll()}function r(l){a(),n.toggle(t.region,l)}function o(l){n.runItem(t.region,l)}return(l,d)=>s.value.length?(k(),T("div",{key:0,class:"dynamic-menu-root","data-dynamic-menu-region":e.region},[(k(!0),T(Q,null,Se(s.value,(c,h)=>(k(),T("div",{key:c.key,class:R(["dropdown",{open:g(n).isOpen(e.region,h)}])},[f("span",{class:"panel-menu","data-dropdown":c.contentId,onClick:it(m=>r(h),["stop","prevent"])},ie(c.label),9,Cc),f("div",{id:c.contentId,class:"dropdown-content"},[(k(!0),T(Q,null,Se(c.items,m=>(k(),T(Q,{key:m.key},[m.type==="separator"?(k(),T("div",Ac)):(k(),T("a",{key:1,href:"#",class:R({disabled:m.enabled===!1}),"aria-disabled":m.enabled===!1?"true":void 0,"data-menu-action":m.action||"","data-menu-dispatch":m.dispatch||"",onClick:it(E=>o(m),["stop","prevent"])},ie(m.label),11,Dc))],64))),128))],8,kc)],2))),128))],8,Ic)):dt("",!0)}},Tc={class:"panel-header"},Lc={class:"panel-header-actions"},Pc={id:"arg-inv-menu",class:"dropdown-content"},Nc={__name:"ArgPane",setup(e){const t=xt(),n=qe();function i(s){return{"menu-flash":t.flashingItemId===s}}return(s,a)=>(k(),T("div",{id:"arg-panel",class:"panel",style:Te(g(n).argPanelStyle)},[f("div",Tc,[a[9]||(a[9]=f("strong",{class:"column-title"},"ARG (Abstract Reachability Graph)",-1)),f("div",Lc,[f("div",{class:R(["dropdown",{open:g(t).isOpen("arg-inv-menu")}])},[f("span",{class:"panel-menu","data-dropdown":"arg-inv-menu",onClick:a[0]||(a[0]=r=>g(Gn)(r,g(t),"arg-inv-menu"))},"Invariant"),f("div",Pc,[f("a",{href:"#",id:"arg-check-induction",class:R(i("arg-check-induction")),onClick:a[1]||(a[1]=r=>g(X)(r,()=>g(L)("checkInduction")))},"Check induction",2),f("a",{href:"#",id:"arg-bounded-check",class:R(i("arg-bounded-check")),onClick:a[2]||(a[2]=r=>g(X)(r,()=>g(L)("boundedCheck")))},"Bounded check",2),f("a",{href:"#",id:"arg-diagram",class:R(i("arg-diagram")),onClick:a[3]||(a[3]=r=>g(X)(r,()=>g(L)("diagramDomain")))},"Diagram",2),f("a",{href:"#",id:"arg-weaken",class:R(i("arg-weaken")),onClick:a[4]||(a[4]=r=>g(X)(r,()=>g(L)("weakenInvariant")))},"Weaken",2),a[8]||(a[8]=f("div",{class:"dropdown-sep"},null,-1)),f("a",{href:"#",id:"arg-save-invariant",class:R(i("arg-save-invariant")),onClick:a[5]||(a[5]=r=>g(X)(r,()=>g(L)("saveInvariant")))},"Save Invariant...",2),f("a",{href:"#",id:"arg-save-abs",class:R(i("arg-save-abs")),onClick:a[6]||(a[6]=r=>g(X)(r,()=>g(L)("saveAbstraction")))},"Save Abstraction...",2)])],2),q(er,{region:"arg"})])]),f("div",{id:"arg-graph",class:"graph-container",onContextmenu:a[7]||(a[7]=it(()=>{},["prevent"]))},null,32)],4))}},Bc={id:"concept-panel",class:"panel"},Fc={class:"panel-header"},Rc={class:"panel-header-actions"},Mc={id:"conj-menu",class:"dropdown-content"},Vc={id:"view-menu",class:"dropdown-content"},Gc={__name:"ConceptPane",setup(e){const t=xt();function n(i){return{"menu-flash":t.flashingItemId===i}}return(i,s)=>(k(),T("div",Bc,[f("div",Fc,[s[25]||(s[25]=f("strong",{class:"column-title"},"Concept graph",-1)),f("div",Rc,[f("div",{class:R(["dropdown",{open:g(t).isOpen("conj-menu")}])},[f("span",{class:"panel-menu","data-dropdown":"conj-menu",onClick:s[0]||(s[0]=a=>g(Gn)(a,g(t),"conj-menu"))},"Conjecture"),f("div",Mc,[f("a",{href:"#",id:"conj-undo",class:R(n("conj-undo")),onClick:s[1]||(s[1]=a=>g(X)(a,()=>g(L)("doUndo")))},"Undo",2),f("a",{href:"#",id:"conj-redo",class:R(n("conj-redo")),onClick:s[2]||(s[2]=a=>g(X)(a,()=>g(L)("doRedo")))},"Redo",2),s[23]||(s[23]=f("div",{class:"dropdown-sep"},null,-1)),f("a",{href:"#",id:"conj-pdr-step",class:R(n("conj-pdr-step")),onClick:s[3]||(s[3]=a=>g(X)(a,()=>g(L)("pdrStep")))},"PDR step",2),f("a",{href:"#",id:"conj-concrete",class:R(n("conj-concrete")),onClick:s[4]||(s[4]=a=>g(X)(a,()=>g(L)("concreteStep")))},"Concrete",2),f("a",{href:"#",id:"conj-gather",class:R(n("conj-gather")),onClick:s[5]||(s[5]=a=>g(X)(a,()=>g(L)("gatherFacts")))},"Gather",2),f("a",{href:"#",id:"conj-cti-gather",class:R(n("conj-cti-gather")),onClick:s[6]||(s[6]=a=>g(X)(a,()=>g(L)("ctiConceptAction","cti_gather")))},"CTI Gather",2),f("a",{href:"#",id:"conj-cti-minimize",class:R(n("conj-cti-minimize")),onClick:s[7]||(s[7]=a=>g(X)(a,()=>g(L)("ctiConceptAction","cti_minimize")))},"Minimize",2),f("a",{href:"#",id:"conj-cti-check-sufficient",class:R(n("conj-cti-check-sufficient")),onClick:s[8]||(s[8]=a=>g(X)(a,()=>g(L)("ctiConceptAction","cti_check_sufficient")))},"Check sufficient",2),f("a",{href:"#",id:"conj-cti-check-inductive",class:R(n("conj-cti-check-inductive")),onClick:s[9]||(s[9]=a=>g(X)(a,()=>g(L)("ctiConceptAction","cti_check_inductive")))},"Check relative induction",2),f("a",{href:"#",id:"conj-cti-strengthen",class:R(n("conj-cti-strengthen")),onClick:s[10]||(s[10]=a=>g(X)(a,()=>g(L)("ctiConceptAction","cti_strengthen")))},"Strengthen",2),f("a",{href:"#",id:"conj-reverse",class:R(n("conj-reverse")),onClick:s[11]||(s[11]=a=>g(X)(a,()=>g(L)("reverseStep")))},"Reverse",2),f("a",{href:"#",id:"conj-path-reach",class:R(n("conj-path-reach")),onClick:s[12]||(s[12]=a=>g(X)(a,()=>g(L)("pathReach")))},"Path reach",2),f("a",{href:"#",id:"conj-reach",class:R(n("conj-reach")),onClick:s[13]||(s[13]=a=>g(X)(a,()=>g(L)("reachStep")))},"Reach",2),f("a",{href:"#",id:"conj-conjecture",class:R(n("conj-conjecture")),onClick:s[14]||(s[14]=a=>g(X)(a,()=>g(L)("makeConjecture")))},"Conjecture",2),f("a",{href:"#",id:"conj-backtrack",class:R(n("conj-backtrack")),onClick:s[15]||(s[15]=a=>g(X)(a,()=>g(L)("backtrack")))},"Backtrack",2),s[24]||(s[24]=f("div",{class:"dropdown-sep"},null,-1)),f("a",{href:"#",id:"conj-recalculate",class:R(n("conj-recalculate")),onClick:s[16]||(s[16]=a=>g(X)(a,()=>g(L)("recalculateGraph")))},"Recalculate",2),f("a",{href:"#",id:"conj-diagram",class:R(n("conj-diagram")),onClick:s[17]||(s[17]=a=>g(X)(a,()=>g(L)("diagramDomain")))},"Diagram",2),f("a",{href:"#",id:"conj-remember",class:R(n("conj-remember")),onClick:s[18]||(s[18]=a=>g(X)(a,()=>g(L)("rememberGraph")))},"Remember",2),f("a",{href:"#",id:"conj-export",class:R(n("conj-export")),onClick:s[19]||(s[19]=a=>g(X)(a,()=>g(L)("exportConjecture")))},"Export",2)])],2),f("div",{class:R(["dropdown",{open:g(t).isOpen("view-menu")}])},[f("span",{class:"panel-menu","data-dropdown":"view-menu",onClick:s[20]||(s[20]=a=>g(Gn)(a,g(t),"view-menu"))},"View"),f("div",Vc,[f("a",{href:"#",id:"view-add-relation",class:R(n("view-add-relation")),onClick:s[21]||(s[21]=a=>g(X)(a,()=>g(L)("addRelationFromString")))},"Add relation",2)])],2),q(er,{region:"concept"})])]),f("div",{id:"concept-graph",class:"graph-container",onContextmenu:s[22]||(s[22]=it(()=>{},["prevent"]))},null,32)]))}},xn="Select a node or edge to see details",tr=Ee("details",{state:()=>({text:xn,facts:[],factCallback:null,traceActionVisible:!1,traceActionCallback:null}),actions:{setDetails({shortInfo:e="",longInfo:t=""}={}){let n=[];e&&n.push(e),t&&(Array.isArray(t)?n=n.concat(t):n.push(t)),this.text=n.join(`
`)||xn,this.facts=[],this.factCallback=null,this.traceActionVisible=!1,this.traceActionCallback=null},clear(){this.text=xn,this.facts=[],this.factCallback=null,this.traceActionVisible=!1,this.traceActionCallback=null},setConstraintFacts(e=[],t=null){this.facts=e.map((n,i)=>({index:typeof n.index=="number"?n.index:i,text:n.text||"",selected:n.selected!==!1})),this.factCallback=typeof t=="function"?t:null,this.text=this.facts.length>0?"":xn,this.traceActionVisible=!1,this.traceActionCallback=null},async toggleFact(e){const t=this.facts.find(i=>i.index===e);if(!t)return;const n=t.selected;if(t.selected=!n,!!this.factCallback)try{await this.factCallback(e,t.selected)}catch(i){throw t.selected=n,i}},setTraceAction(e){this.traceActionCallback=typeof e=="function"?e:null,this.traceActionVisible=!!this.traceActionCallback},runTraceAction(){this.traceActionCallback&&this.traceActionCallback()}}}),Hc={id:"info-content"},Oc=["data-constraint-fact","aria-pressed","onClick"],jc={__name:"DetailsPane",setup(e){const t=tr(),n=qe();return(i,s)=>(k(),T("div",{id:"info-panel",class:"info-panel",style:Te(g(n).detailsPanelStyle)},[s[3]||(s[3]=f("div",{id:"info-header",class:"info-header",title:"Drag to resize Details"},"Details",-1)),f("div",Hc,[g(t).facts.length>0?(k(),T(Q,{key:0},[s[1]||(s[1]=f("div",{class:"constraint-facts-title"},"Constraints:",-1)),(k(!0),T(Q,null,Se(g(t).facts,a=>(k(),T("button",{key:a.index,type:"button",class:R(["constraint-fact",{inactive:!a.selected}]),"data-constraint-fact":String(a.index),"aria-pressed":a.selected?"true":"false",onClick:r=>g(t).toggleFact(a.index)},ie(a.text),11,Oc))),128))],64)):(k(),T(Q,{key:1},[sn(ie(g(t).text),1)],64)),g(t).traceActionVisible?(k(),T(Q,{key:2},[s[2]||(s[2]=f("br",null,null,-1)),f("button",{type:"button",class:"btn small","data-check-view-trace":"true",onClick:s[0]||(s[0]=a=>g(t).runTraceAction())},"View error trace")],64)):dt("",!0)])],4))}};function nr(e=[],t=""){return(e||[]).map((n,i)=>{const s=n&&n.address!=null?String(n.address):t===""?String(i):`${t}/${i}`;return{...n||{},address:s,text:n&&n.text||"",subs:nr(n&&n.subs||[],s)}})}function $c(e){const t=String(e||"").split("/"),n=[];for(let i=0;i<t.length-1;i+=1)n.push(t.slice(0,i+1).join("/"));return n}const ss=Ee("eventTrace",{state:()=>({sheets:{},expandedBySheet:{},selectedPatternIndexBySheet:{}}),getters:{sheetList:e=>Object.values(e.sheets),sheetById:e=>t=>e.sheets[t]||null,isExpanded:e=>(t,n)=>!!(e.expandedBySheet[t]&&e.expandedBySheet[t][n]),selectedPatternIndex:e=>t=>e.selectedPatternIndexBySheet[t]??-1,selectedPattern:e=>t=>{const n=e.sheets[t],i=e.selectedPatternIndexBySheet[t]??-1;return n&&i>=0&&n.patterns[i]||""}},actions:{upsertSheet(e={}){if(!e.id)return;const t=String(e.id);this.sheets[t]={id:t,label:e.label||t,events:nr(e.events||[]),patterns:Array.isArray(e.patterns)?e.patterns.slice():[],selectedEventAddress:e.selectedEventAddress||e.selected_address||""},this.expandedBySheet[t]||(this.expandedBySheet[t]={}),this.sheets[t].selectedEventAddress&&this.selectEvent(t,this.sheets[t].selectedEventAddress)},removeSheet(e){delete this.sheets[e],delete this.expandedBySheet[e],delete this.selectedPatternIndexBySheet[e]},reset(){this.sheets={},this.expandedBySheet={},this.selectedPatternIndexBySheet={}},setExpanded(e,t,n){this.expandedBySheet[e]||(this.expandedBySheet[e]={}),this.expandedBySheet[e][t]=!!n},selectEvent(e,t){const n=this.sheets[e];n&&(n.selectedEventAddress=t||"",$c(t).forEach(i=>this.setExpanded(e,i,!0)))},setPatterns(e,t=[]){const n=this.sheets[e];if(!n)return;n.patterns=Array.isArray(t)?t.slice():[],(this.selectedPatternIndexBySheet[e]??-1)>=n.patterns.length&&(this.selectedPatternIndexBySheet[e]=-1)},setSelectedPatternIndex(e,t){this.selectedPatternIndexBySheet[e]=Number.isInteger(t)?t:-1}}}),vt={id:"sheet-1",label:"Sheet 1",closable:!1,type:"analysis"},is=Ee("sheets",{state:()=>({tabs:[vt],activeSheetId:vt.id}),getters:{labelFor:e=>t=>{const n=e.tabs.find(i=>i.id===t);return n?n.label:""}},actions:{upsertTab(e){if(!e||!e.id)return;const t={id:String(e.id),label:e.label||e.id,closable:e.closable!==!1&&e.id!==vt.id,type:e.type||"analysis"},n=this.tabs.findIndex(i=>i.id===t.id);n>=0?this.tabs[n]={...this.tabs[n],...t}:this.tabs.push(t)},activateTab(e){e&&(this.activeSheetId=String(e))},removeTab(e){!e||e===vt.id||(this.tabs=this.tabs.filter(t=>t.id!==e),this.activeSheetId===e&&(this.activeSheetId=vt.id))},resetTabs(){this.tabs=[vt],this.activeSheetId=vt.id}}}),zc=["data-event-node"],Wc=["data-event-address"],Uc=["disabled","data-event-toggle"],Kc={class:"event-text"},qc={key:0,class:"event-tree-list"},Yc={__name:"EventTraceNode",props:{sheetId:{type:String,required:!0},event:{type:Object,required:!0}},setup(e){const t=e,n=ss(),i=Ne(()=>t.event.subs&&t.event.subs.length>0),s=Ne(()=>n.isExpanded(t.sheetId,t.event.address)),a=Ne(()=>{const l=n.sheetById(t.sheetId);return!!l&&l.selectedEventAddress===t.event.address});function r(){if(pn("selectEventTraceRow")){L("selectEventTraceRow",t.sheetId,t.event.address);return}n.selectEvent(t.sheetId,t.event.address)}function o(l){if(l.stopPropagation(),!!i.value){if(pn("toggleEventTraceNode")){L("toggleEventTraceNode",t.sheetId,t.event.address);return}n.setExpanded(t.sheetId,t.event.address,!s.value)}}return(l,d)=>{const c=wo("EventTraceNode",!0);return k(),T("li",{class:"event-tree-node","data-event-node":e.event.address},[f("div",{class:R(["event-row",{selected:a.value}]),"data-event-address":e.event.address,onClick:r},[f("button",{class:"event-toggle",type:"button",disabled:!i.value,"data-event-toggle":i.value?e.event.address:void 0,onClick:o},ie(i.value?s.value?"-":"+":""),9,Uc),f("span",Kc,ie(e.event.text),1)],10,Wc),i.value&&s.value?(k(),T("ul",qc,[(k(!0),T(Q,null,Se(e.event.subs,h=>(k(),qs(c,{key:h.address,"sheet-id":e.sheetId,event:h},null,8,["sheet-id","event"]))),128))])):dt("",!0)],8,zc)}}},Xc={class:"event-viewer"},Jc={class:"event-tree-panel panel"},Zc={class:"panel-header"},Qc=["data-event-tree"],ed={class:"event-tree-list"},td={class:"event-pattern-panel panel"},nd=["value"],sd=["value"],id={class:"event-pattern-buttons"},ad={class:"event-pattern-buttons"},rd={__name:"EventTraceSheet",props:{sheetId:{type:String,required:!0}},setup(e){const t=e,n=ss(),i=Ne(()=>n.sheetById(t.sheetId));async function s(l,d,c,h){if(!pn("entryDialog"))return;const m=await L("entryDialog",l,d,"",{okLabel:c});m!==null&&m!==""&&await h(m)}function a(){return n.selectedPattern(t.sheetId)}function r(l){return l.target.selectedIndex}async function o(){if(!pn("textDialog"))return;const l=await L("textDialog","Load patterns","Paste patterns:","",{okLabel:"Load"});l!==null&&await L("loadEventPatterns",t.sheetId,l)}return(l,d)=>(k(),T("div",Xc,[f("div",Jc,[f("div",Zc,[d[10]||(d[10]=f("span",{class:"panel-title"},"Events",-1)),f("button",{class:"menu-btn event-filter-btn",type:"button",onClick:d[0]||(d[0]=c=>s("Filter events","Pattern:","Filter",h=>g(L)("filterEventTrace",h)))},"Filter..."),f("button",{class:"menu-btn event-find-fwd-btn",type:"button",onClick:d[1]||(d[1]=c=>s("Find event","Pattern:","Find",h=>g(L)("findEventTrace",h,!1)))},">>"),f("button",{class:"menu-btn event-find-rev-btn",type:"button",onClick:d[2]||(d[2]=c=>s("Find event","Pattern:","Find",h=>g(L)("findEventTrace",h,!0)))},"<<")]),f("div",{class:"event-tree","data-event-tree":e.sheetId},[f("ul",ed,[(k(!0),T(Q,null,Se(i.value&&i.value.events||[],c=>(k(),qs(Yc,{key:c.address,"sheet-id":e.sheetId,event:c},null,8,["sheet-id","event"]))),128))])],8,Qc)]),f("div",td,[d[11]||(d[11]=f("div",{class:"panel-header"},[f("span",{class:"panel-title"},"Patterns")],-1)),f("select",{class:"event-pattern-list",size:"8",value:a(),onChange:d[3]||(d[3]=c=>g(n).setSelectedPatternIndex(e.sheetId,r(c)))},[(k(!0),T(Q,null,Se(i.value&&i.value.patterns||[],c=>(k(),T("option",{key:c,value:c},ie(c),9,sd))),128))],40,nd),f("div",id,[f("button",{class:"menu-btn event-pattern-rev",type:"button",onClick:d[4]||(d[4]=c=>a()&&g(L)("findEventTrace",a(),!0))},"<<"),f("button",{class:"menu-btn event-pattern-fwd",type:"button",onClick:d[5]||(d[5]=c=>a()&&g(L)("findEventTrace",a(),!1))},">>"),f("button",{class:"menu-btn event-pattern-add",type:"button",onClick:d[6]||(d[6]=c=>s("Add pattern","Pattern:","Add",h=>g(L)("addEventPattern",e.sheetId,h)))},"+"),f("button",{class:"menu-btn event-pattern-remove",type:"button",onClick:d[7]||(d[7]=c=>g(L)("removeSelectedEventPattern",e.sheetId))},"-")]),f("div",ad,[f("button",{class:"menu-btn event-pattern-save",type:"button",onClick:d[8]||(d[8]=c=>g(L)("saveEventPatterns",e.sheetId))},"Save"),f("button",{class:"menu-btn event-pattern-load",type:"button",onClick:o},"Load"),f("button",{class:"menu-btn event-pattern-clear",type:"button",onClick:d[9]||(d[9]=c=>g(L)("clearEventPatterns",e.sheetId))},"Clear")])])]))}},od=["id"],ld={__name:"EventTraceSheetHost",setup(e){const t=ss(),n=is();return(i,s)=>(k(!0),T(Q,null,Se(g(t).sheetList,a=>(k(),T("div",{id:a.id,key:a.id,class:R(["sheet-content event-sheet",{active:g(n).activeSheetId===a.id}])},[q(rd,{"sheet-id":a.id},null,8,["sheet-id"])],10,od))),128))}},cd={all_to_all:!1,edge_unknown:!1,none_to_none:!1,transitive:!1},ys=["all_to_all","edge_unknown","none_to_none","transitive"],dd=[["all_to_all","node_necessarily"],["edge_unknown","node_maybe"],["none_to_none","node_necessarily_not"]],sr=Ee("stateRelations",{state:()=>({stateLabel:"0",loaded:!1,rows:[],toggleCallback:null}),getters:{hasRows:e=>e.rows.length>0,showPlaceholder:e=>e.loaded&&e.rows.length===0,toggleSnapshot:e=>{const t={};return e.rows.forEach(n=>{ys.forEach(i=>{t[`${n.name}|${i}`]=!!n.checked[i]})}),t},visibilitySnapshot:e=>{const t={},n={};return e.rows.forEach(i=>{t[i.name]={},ys.forEach(s=>{t[i.name][s]=!!i.checked[s]}),n[i.name]={},dd.forEach(([s,a])=>{n[i.name][a]=!!i.checked[s]})}),{edges:t,labels:n}}},actions:{setStateLabel(e){this.stateLabel=e==null?"—":String(e)},setRows(e=[],t=null){this.loaded=!0,this.rows=e.map(n=>({name:n.name,checked:{...cd,...n.checked||{}}})),this.toggleCallback=typeof t=="function"?t:null},clear(){this.loaded=!1,this.rows=[],this.toggleCallback=null,this.stateLabel="0"},applyToggleSnapshot(e={}){this.rows.forEach(t=>{ys.forEach(n=>{const i=`${t.name}|${n}`;Object.prototype.hasOwnProperty.call(e,i)&&(t.checked[n]=!!e[i])})})},toggle(e,t,n){const i=this.rows.find(s=>s.name===e);i&&(i.checked[t]=n),this.toggleCallback&&this.toggleCallback(e,t,n)}}}),ud={class:"panel-header"},hd={class:"panel-header-actions"},fd={id:"state-label"},pd={id:"state-controls"},gd={id:"state-checkbox-table"},vd={id:"state-checkbox-body"},md=["name","value","title","checked","onChange"],yd={class:"name-col"},bd={key:0},_d={__name:"StateRelationsPane",setup(e){const t=sr(),n=qe(),i=[{key:"all_to_all",label:"+",title:s=>`Show definite edges (${s})`},{key:"edge_unknown",label:"?",title:s=>`Show unknown edges (${s})`},{key:"none_to_none",label:"-",title:s=>`Show absent edges (${s})`},{key:"transitive",label:"T",title:s=>`Transitive reduction (${s})`}];return(s,a)=>(k(),T("div",{id:"state-panel",class:"panel",style:Te(g(n).statePanelStyle)},[f("div",ud,[a[1]||(a[1]=f("strong",{class:"column-title"},"State/relations",-1)),f("div",hd,[f("span",fd,"State: "+ie(g(t).stateLabel),1)])]),f("div",pd,[f("table",gd,[a[3]||(a[3]=f("thead",null,[f("tr",null,[f("th",{class:"chk-col"},"+"),f("th",{class:"chk-col"},"?"),f("th",{class:"chk-col"},"-"),f("th",{class:"chk-col"},"T"),f("th",{class:"name-col"})])],-1)),f("tbody",vd,[(k(!0),T(Q,null,Se(g(t).rows,r=>(k(),T("tr",{key:r.name},[(k(),T(Q,null,Se(i,o=>f("td",{key:o.key},[f("input",{type:"checkbox",name:r.name,value:o.key,title:o.title(r.name),checked:r.checked[o.key],onChange:l=>g(t).toggle(r.name,o.key,l.target.checked)},null,40,md)])),64)),f("td",yd,[f("a",{href:"#",onClick:a[0]||(a[0]=it(()=>{},["prevent"]))},ie(r.name),1)])]))),128)),g(t).showPlaceholder?(k(),T("tr",bd,[...a[2]||(a[2]=[f("td",{colspan:"5",class:"state-relations-placeholder"},"No relations loaded",-1)])])):dt("",!0)])])])],4))}},wd={id:"tab-bar"},Sd=["data-sheet","onClick"],Ed=["onClick"],xd={__name:"TabBar",setup(e){const t=is();return(n,i)=>(k(),T("div",wd,[(k(!0),T(Q,null,Se(g(t).tabs,s=>(k(),T("button",{key:s.id,class:R(["sheet-tab",{active:g(t).activeSheetId===s.id}]),"data-sheet":s.id,onClick:a=>g(ge)(a,()=>g(L)("switchSheet",s.id))},[f("span",null,ie(s.label),1),s.closable?(k(),T("span",{key:0,class:"tab-close",title:"Close tab",onClick:a=>g(ge)(a,()=>g(L)("removeSheet",s.id))},"×",8,Ed)):dt("",!0)],10,Sd))),128))]))}};function Id(){return globalThis.window&&globalThis.window.ivyApp}function bs(){const e=Id();if(e){if(typeof e._refreshGraphsAndEditorLayout=="function"){e._refreshGraphsAndEditorLayout();return}e.argGraph&&typeof e.argGraph.resize=="function"&&e.argGraph.resize(),e.conceptGraph&&typeof e.conceptGraph.resize=="function"&&e.conceptGraph.resize()}}function ze(){bs(),Et(bs);const e=globalThis.window;e&&typeof e.requestAnimationFrame=="function"&&e.requestAnimationFrame(bs)}function In(e,t){!e||typeof e.querySelectorAll!="function"||e.querySelectorAll(".graph-container").forEach(n=>{n.style.pointerEvents=t})}function rn(e,{cursor:t,onMove:n,onEnd:i,onStart:s}={}){const a=e.currentTarget?e.currentTarget.ownerDocument:globalThis.document;if(!a)return;typeof e.preventDefault=="function"&&e.preventDefault(),typeof s=="function"&&s(),a.body.style.cursor=t||"",a.body.style.userSelect="none";const r=l=>{typeof n=="function"&&n(l)},o=l=>{a.removeEventListener("mousemove",r),a.removeEventListener("mouseup",o),a.body.style.cursor="",a.body.style.userSelect="",typeof i=="function"&&i(l)};a.addEventListener("mousemove",r),a.addEventListener("mouseup",o)}const Cd={class:"sheet-columns"},kd={class:"sheet-left"},Ad={class:"sheet-main"},Dd={__name:"SheetArea",setup(e){const t=is(),n=qe();function i(l){if(!l)return 44;const c=l.querySelector(".sheet-main");if(!c)return 44;let h=44;return c.querySelectorAll(".panel-header").forEach(m=>{h=Math.max(h,m.offsetHeight||0)}),h}function s(l,d){const c=d.parentElement,h=d.previousElementSibling;if(!c||!h)return;const m=l.clientX,E=h.offsetWidth;d.classList.add("active"),In(l.currentTarget,"none"),rn(l,{cursor:"col-resize",onMove(B){const D=c.offsetWidth||800,W=Math.max(150,Math.min(E+B.clientX-m,D-200));h.id==="arg-panel"?n.setArgPanelWidth(W):h.style.flex=`0 0 ${W}px`,ze()},onEnd(){d.classList.remove("active"),In(l.currentTarget,""),ze()}})}function a(l,d){const c=document.getElementById("state-panel"),h=d.parentElement;if(!c)return;const m=l.clientX,E=c.offsetWidth;d.classList.add("active"),rn(l,{cursor:"col-resize",onMove(B){const D=h?h.offsetWidth-300:800,W=Math.max(200,Math.min(E+m-B.clientX,D));n.setStatePanelWidth(W),ze()},onEnd(){d.classList.remove("active"),ze()}})}function r(l,d){const c=d.closest(".info-panel")||d.parentElement,h=c?c.closest(".sheet-left"):null;if(!c||!h)return;const m=h.querySelector(".sheet-main"),E=l.clientY,B=c.offsetHeight;d.classList.add("active"),In(h,"none"),rn(l,{cursor:"row-resize",onMove(D){const U=i(h),V=Math.max(72,h.offsetHeight-U),H=Math.max(72,Math.min(B+E-D.clientY,V));m&&(m.style.minHeight=`${U}px`),c.id==="info-panel"?n.setDetailsHeight(H):(c.style.flex=`0 0 ${H}px`,c.style.height=`${H}px`),ze()},onEnd(){d.classList.remove("active"),In(h,""),ze()}})}function o(l){const d=l.target;if(!(d instanceof Element))return;if(d.id==="divider2"){a(l,d);return}const c=d.closest(".info-header, #info-header");if(c&&l.currentTarget.contains(c)){r(l,c);return}d.classList.contains("divider")&&d.parentElement&&d.parentElement.classList.contains("sheet-main")&&s(l,d)}return(l,d)=>(k(),T("div",{id:"sheet-area",onMousedown:o},[q(xd),f("div",{id:"sheet-1",class:R(["sheet-content",{active:g(t).activeSheetId==="sheet-1"}])},[f("div",Cd,[f("div",kd,[f("div",Ad,[q(Nc),d[0]||(d[0]=f("div",{id:"divider",class:"divider"},null,-1)),q(Gc)]),q(jc)]),d[1]||(d[1]=f("div",{id:"divider2",class:"divider"},null,-1)),q(_d)])],2),q(ld)],32))}},Td={class:"panel-header tutorial-url-bar"},Ld=["disabled"],Pd=["disabled"],Nd=["value"],Bd=["src"],Fd={__name:"TutorialPane",setup(e){const t=qe();function n(a){a.key==="Enter"&&ge(a,()=>t.navigateTutorial(t.tutorialInput))}function i(){if(pn("toggleTutorial")){L("toggleTutorial",!0);return}t.setTutorialVisible(!1)}function s(a){try{const r=a.target.contentWindow.location.href;t.recordTutorialLoad(r)}catch{}}return(a,r)=>(k(),T("div",{id:"tutorial-container",style:Te(g(t).tutorialPanelStyle)},[f("div",Td,[f("button",{id:"tutorial-back",class:"tutorial-nav-btn",title:"Back",disabled:!g(t).canGoBack,onClick:r[0]||(r[0]=o=>g(ge)(o,()=>g(t).goTutorialBack()))},"◀",8,Ld),f("button",{id:"tutorial-fwd",class:"tutorial-nav-btn",title:"Forward",disabled:!g(t).canGoForward,onClick:r[1]||(r[1]=o=>g(ge)(o,()=>g(t).goTutorialForward()))},"▶",8,Pd),f("button",{id:"tutorial-reload",class:"tutorial-nav-btn",title:"Reload",onClick:r[2]||(r[2]=o=>g(ge)(o,()=>g(t).reloadTutorial()))},"↻"),f("input",{id:"tutorial-url",type:"text",class:"tutorial-url-input",value:g(t).tutorialInput,spellcheck:"false",onInput:r[3]||(r[3]=o=>g(t).setTutorialInput(o.target.value)),onKeydown:n},null,40,Nd),f("button",{id:"tutorial-close",class:"tutorial-close-btn",title:"Close tutorial",onClick:r[4]||(r[4]=o=>g(ge)(o,i))},"×")]),(k(),T("iframe",{id:"tutorial-iframe",key:g(t).tutorialFrameKey,class:"tutorial-iframe",src:g(t).tutorialUrl,onLoad:s},null,40,Bd))],4))}},Rd={id:"outer-container"},Md={id:"top-row"},Vd={__name:"WorkspaceShell",setup(e){const t=qe();function n(s){const a=s.currentTarget,r=document.getElementById("sheet-area"),o=document.getElementById("editor-panel"),l=document.getElementById("top-row");if(!a||!r||!o||!l)return;const d=s.clientX,c=o.offsetWidth,h=200,m=400;r.style.flex="1 1 auto",a.classList.add("active"),rn(s,{cursor:"col-resize",onMove(E){const B=a.offsetWidth||4,D=Math.max(h,l.offsetWidth-B-m),W=Math.max(h,Math.min(c+d-E.clientX,D));t.setEditorWidth(W),ze()},onEnd(){a.classList.remove("active"),ze()}})}function i(s){const a=s.currentTarget,r=document.getElementById("tutorial-container"),o=document.getElementById("outer-container"),l=document.getElementById("tutorial-iframe");if(!a||!r)return;const d=s.clientY,c=r.offsetHeight;a.classList.add("active"),l&&(l.style.pointerEvents="none"),rn(s,{cursor:"row-resize",onMove(h){const m=o?o.offsetHeight-100:600,E=Math.max(80,Math.min(c+d-h.clientY,m));t.setTutorialHeight(E),ze()},onEnd(){a.classList.remove("active"),l&&(l.style.pointerEvents=""),ze()}})}return(s,a)=>(k(),T("div",Rd,[f("div",Md,[q(Dd),f("div",{id:"divider3",class:"divider",onMousedown:n},null,32),q(Sc)]),dn(f("div",{id:"divider-h",class:"divider-horizontal",onMousedown:i},null,544),[[fn,g(t).tutorialVisible]]),q(Fd)]))}},Js=Ee("contextMenu",{state:()=>({visible:!1,x:0,y:0,items:[]}),getters:{style:e=>({display:e.visible?"block":"none",left:`${e.x}px`,top:`${e.y}px`})},actions:{show(e,t,n=[]){this.x=Number(e)||0,this.y=Number(t)||0,this.items=n.map((i,s)=>i.separator?{kind:"separator",key:`sep-${s}`}:i.header?{kind:"header",key:`hdr-${s}`,header:i.header}:{kind:"item",key:`item-${s}-${i.id||i.name||""}`,name:i.name||"",id:i.id||i.name||"",callback:typeof i.callback=="function"?i.callback:null}),this.visible=!0},setPosition(e,t){this.x=Number(e)||0,this.y=Number(t)||0},hide(){this.visible=!1,this.items=[]},runItem(e){this.hide(),e&&typeof e.callback=="function"&&e.callback()}}}),Gd={key:0,class:"context-menu-separator"},Hd={key:1,class:"context-menu-header"},Od=["data-action-id","onClick"],jd={__name:"ContextMenuHost",setup(e){const t=Js(),n=Jt(null);function i(){if(!t.visible||!n.value)return;const s=n.value.getBoundingClientRect(),a=window.innerWidth||0,r=window.innerHeight||0,o=Math.max(0,a-s.width),l=Math.max(0,r-s.height),d=Math.min(t.x,o),c=Math.min(t.y,l);(d!==t.x||c!==t.y)&&t.setPosition(d,c)}return Nt(()=>[t.visible,t.x,t.y,t.items.length],async()=>{await Et(),i()},{flush:"post",immediate:!0}),(s,a)=>(k(),T("div",{id:"context-menu",ref_key:"menuEl",ref:n,class:"context-menu",style:Te(g(t).style)},[(k(!0),T(Q,null,Se(g(t).items,r=>(k(),T(Q,{key:r.key},[r.kind==="separator"?(k(),T("div",Gd)):r.kind==="header"?(k(),T("div",Hd,ie(r.header),1)):(k(),T("div",{key:2,class:"context-menu-item","data-action-id":r.id,onClick:it(o=>g(t).runItem(r),["stop"])},ie(r.name),9,Od))],64))),128))],4))}};function $d(e=[]){return e.map(t=>typeof t=="object"&&t!==null?{label:t.label||String(t.value),value:t.value}:{label:String(t),value:t})}const ir=Ee("dialogs",{state:()=>({active:null}),actions:{open(e={}){return new Promise(t=>{const n=e.type||"ok",i=n==="listbox"?$d(e.items||[]):[];this.active={...e,type:n,title:e.title||"ivyweb",message:e.message||"",options:e.options||{},entries:i,inputValue:String(e.initialValue==null?e.text||"":e.initialValue),selectedIndex:"",selectedIndices:[],error:"",resolve:t}})},clearError(){this.active&&(this.active.error="")},setInputValue(e){this.active&&(this.active.inputValue=e,this.clearError())},setSelectedIndex(e){this.active&&(this.active.selectedIndex=e)},setSelectedIndices(e){this.active&&(this.active.selectedIndices=Array.isArray(e)?e:[])},finish(e){const t=this.active;this.active=null,t&&typeof t.resolve=="function"&&t.resolve(e)},cancelValue(){return this.active&&this.active.type==="listbox"&&this.active.options.multiple?[]:null},cancel(){this.finish(this.cancelValue())},escape(){this.active&&(this.active.type==="ok"?this.finish(!0):this.active.type==="okCancel"?this.finish(!1):this.active.escapeValue!==void 0?this.finish(this.active.escapeValue):this.cancel())},submit(){if(!this.active)return;const e=this.active,t=e.options||{};if(e.type==="ok"){this.finish(!0);return}if(e.type==="okCancel"){this.finish(!0);return}if(e.type==="text"||e.type==="entry"){this.finish(e.inputValue);return}if(e.type==="integer"){const n=String(e.inputValue||"").trim(),i=Number(n);if(n===""||!Number.isInteger(i)){e.error="Enter an integer.";return}if(t.min!=null&&i<t.min){e.error=`Enter a value at least ${t.min}.`;return}if(t.max!=null&&i>t.max){e.error=`Enter a value at most ${t.max}.`;return}this.finish(i);return}if(e.type==="listbox"){if(t.multiple){this.finish(e.selectedIndices.map(i=>e.entries[Number(i)]).filter(Boolean).map(i=>i.value));return}const n=e.entries[Number(e.selectedIndex)];this.finish(n?n.value:null)}},chooseButton(e){this.finish(e?e.value:null)}}}),zd={key:0,class:"dialog-overlay","data-ivy-dialog":"true"},Wd={class:"dialog-box"},Ud={class:"dialog-title"},Kd={class:"dialog-message"},qd={class:"dialog-body"},Yd=["rows","cols","readonly","value"],Xd=["value"],Jd=["value","min","max"],Zd=["size","multiple"],Qd=["value","data-ivy-dialog-index"],eu={class:"dialog-buttons"},tu=["onClick"],nu={__name:"DialogHost",setup(e){const t=ir(),n=Jt(null),i=Jt(null),s=Jt(null),a=Ne(()=>t.active),r=Ne(()=>{const d=a.value;if(!d)return[];const c=d.options||{};if(d.type==="ok")return[{label:"OK",action:"submit"}];if(d.type==="okCancel")return[{label:"Cancel",action:"cancel"},{label:"OK",action:"submit"}];if(d.type==="buttons")return(d.buttons||[]).map(m=>({label:m.label||String(m.value),action:"button",value:m.value,danger:!!m.danger}));const h=[];return(c.cancel||d.type!=="text"&&c.cancel!==!1)&&h.push({label:"Cancel",action:"cancel"}),h.push({label:c.okLabel||"OK",action:"submit"}),h});function o(d){if(d.action==="cancel"){t.cancel();return}if(d.action==="button"){t.chooseButton(d);return}t.submit()}function l(d){d.key==="Escape"&&t.active&&t.escape()}return Nt(a,async d=>{if(!d)return;await Et();const c=n.value||i.value||s.value;c&&typeof c.focus=="function"&&(c.focus(),typeof c.select=="function"&&d.type!=="listbox"&&c.select())}),zs(()=>{document.addEventListener("keydown",l)}),ba(()=>{document.removeEventListener("keydown",l)}),(d,c)=>a.value?(k(),T("div",zd,[f("div",Wd,[f("div",Ud,ie(a.value.title),1),f("div",Kd,ie(a.value.message),1),f("div",qd,[a.value.type==="text"?(k(),T("textarea",{key:0,ref_key:"textInput",ref:n,class:"dialog-text",rows:a.value.options.rows||4,cols:a.value.options.cols||80,readonly:!!a.value.options.readOnly,value:a.value.inputValue,"data-ivy-dialog-text":"true",onInput:c[0]||(c[0]=h=>g(t).setInputValue(h.target.value))},null,40,Yd)):a.value.type==="entry"?(k(),T("input",{key:1,ref_key:"scalarInput",ref:i,type:"text",class:"dialog-input",value:a.value.inputValue,"data-ivy-dialog-entry":"true",onInput:c[1]||(c[1]=h=>g(t).setInputValue(h.target.value))},null,40,Xd)):a.value.type==="integer"?(k(),T("input",{key:2,ref_key:"scalarInput",ref:i,type:"number",class:"dialog-input",value:a.value.inputValue,min:a.value.options.min,max:a.value.options.max,"data-ivy-dialog-int":"true",onInput:c[2]||(c[2]=h=>g(t).setInputValue(h.target.value))},null,40,Jd)):a.value.type==="listbox"?(k(),T("select",{key:3,ref_key:"listInput",ref:s,class:"dialog-input dialog-listbox",size:a.value.options.size||Math.min(Math.max(a.value.entries.length,2),12),multiple:!!a.value.options.multiple,"data-ivy-dialog-list":"true",onChange:c[3]||(c[3]=h=>a.value.options.multiple?g(t).setSelectedIndices(Array.from(h.target.selectedOptions).map(m=>m.getAttribute("data-ivy-dialog-index"))):g(t).setSelectedIndex(h.target.selectedOptions[0]?h.target.selectedOptions[0].getAttribute("data-ivy-dialog-index"):""))},[(k(!0),T(Q,null,Se(a.value.entries,(h,m)=>(k(),T("option",{key:m,value:String(h.value),"data-ivy-dialog-index":String(m)},ie(h.label),9,Qd))),128))],40,Zd)):dt("",!0)]),dn(f("div",{class:"dialog-message dialog-error","data-ivy-dialog-error":"true"},ie(a.value.error),513),[[fn,a.value.error]]),f("div",eu,[(k(!0),T(Q,null,Se(r.value,h=>(k(),T("button",{key:h.label,type:"button",class:R(["dialog-btn",{"dialog-btn-danger":h.danger}]),"data-ivy-dialog-button":"true",onClick:m=>o(h)},ie(h.label),11,tu))),128))])])])):dt("",!0)}},su={__name:"FileInputHost",setup(e){async function t(s){const a=s.target,r=a.files&&a.files[0];r&&await L("loadFile",r),a.value=""}async function n(s){const a=s.target,r=a.files&&a.files[0];r&&await L("loadEventTraceFile",r),a.value=""}async function i(s){const a=s.target,r=a.files&&a.files[0];if(r)try{await L("loadAnalysisStateFile",r)}catch(o){Xl(`Load analysis state failed: ${o.message}`,"error")}a.value=""}return(s,a)=>(k(),T(Q,null,[f("input",{type:"file",id:"file-input",class:"hidden-file-input",accept:".ivy",onChange:t},null,32),f("input",{type:"file",id:"event-file-input",class:"hidden-file-input",accept:".iev,.pats,.txt",onChange:n},null,32),f("input",{type:"file",id:"analysis-state-file-input",class:"hidden-file-input",accept:".json,.ivyweb.json",onChange:i},null,32)],64))}},iu={id:"save-as-explain-notice",class:"save-as-explain-notice"},au={id:"loading-overlay"},ru={id:"loading-message"},ou={__name:"SessionOverlayHost",setup(e){const t=ns();return(n,i)=>(k(),T(Q,null,[dn(f("div",iu," The browser security model requires re-confirmation of the save path on disk when IvyWeb cannot locate an IndexedDB cached file handle. ",512),[[fn,g(t).saveAsNoticeVisible]]),dn(f("div",au,[i[0]||(i[0]=f("div",{class:"spinner"},null,-1)),f("div",ru,ie(g(t).loadingMessage),1)],512),[[fn,g(t).loading]])],64))}},lu={class:"status-message"},cu={id:"session-id",class:"session-id"},du={__name:"StatusBar",setup(e){const t=ns();return(n,i)=>(k(),T("div",{id:"statusbar",class:R(g(t).statusLevel)},[f("span",lu,ie(g(t).status),1),f("span",cu,ie(g(t).sessionDisplay),1)],2))}};let uu=1;const ar=Ee("toasts",{state:()=>({items:[]}),actions:{show(e,t="info",n={}){const i=uu++,s={id:i,message:String(e||""),level:t||"info",className:n.className||"",persistent:!!n.persistent};if(this.items.push(s),!s.persistent){const a=Number(n.timeoutMs)||1e4;window.setTimeout(()=>{this.remove(i)},a)}return i},remove(e){this.items=this.items.filter(t=>t.id!==e)},clear(){this.items=[]}}}),hu={class:"ivy-toast-stack","aria-live":"polite"},fu=["onClick"],pu={__name:"ToastHost",setup(e){const t=ar();return(n,i)=>(k(),T("div",hu,[(k(!0),T(Q,null,Se(g(t).items,s=>(k(),T("button",{key:s.id,type:"button",class:R(["ivy-toast",[`ivy-toast-${s.level}`,s.className]]),onClick:a=>g(t).remove(s.id)},ie(s.message),11,fu))),128))]))}};function gu({doc:e=globalThis.document,contextMenuStore:t,dropdownStore:n,menuDescriptorStore:i}={}){if(!e)return()=>{};const s=()=>{t&&t.hide(),n&&n.closeAll(),i&&i.closeAll()},a=o=>{t&&t.hide(),(!o.target||typeof o.target.closest!="function"||!o.target.closest(".dropdown"))&&(n&&n.closeAll(),i&&i.closeAll())},r=o=>{const l=String(o.key||"").toLowerCase();if(o.key==="Escape"){s();return}if((o.ctrlKey||o.metaKey)&&l==="s"){o.preventDefault(),L("save");return}(o.ctrlKey||o.metaKey)&&l==="z"&&(o.preventDefault(),L("doUndo"))};return e.addEventListener("click",a),e.addEventListener("keydown",r),()=>{e.removeEventListener("click",a),e.removeEventListener("keydown",r)}}async function vu(e=globalThis.window){if(await Et(),e&&typeof e.startIvyApp=="function")return e.startIvyApp()}const mu=Ee("graph",{state:()=>({activeSheetId:"sheet-1",argBySheet:{},conceptBySheet:{},selectedArgNode:null,selectedArgNodeBySheet:{},toggles:{}}),actions:{setActiveSheet(e){e&&(this.activeSheetId=String(e))},applyGraphSnapshot(e,t,n={}){if(!e||t!=="arg"&&t!=="concept")return;const i=t==="arg"?"argBySheet":"conceptBySheet";this[i]={...this[i],[e]:{elements:Array.isArray(n.elements)?n.elements:[],positions:n.positions||null,updatedAt:n.updatedAt||Date.now()}}},async refreshArg(e=this.activeSheetId){const t=await wt().engine.getArg({sheetId:e});return this.argBySheet={...this.argBySheet,[e]:t},t},async refreshConcept({sheetId:e=this.activeSheetId,stateId:t}={}){const n=await wt().engine.getConcept({sheetId:e,stateId:t});return this.conceptBySheet={...this.conceptBySheet,[e]:n},n},async refreshToggles(){return this.toggles=await wt().engine.getToggles(),this.toggles},selectArgNode(e,t=this.activeSheetId){this.selectedArgNode=e,t&&(this.selectedArgNodeBySheet={...this.selectedArgNodeBySheet,[t]:e})}}}),yu={__name:"App",setup(e){const t=Js(),n=xt(),i=Xs();let s=null;return zs(()=>{s=gu({contextMenuStore:t,dropdownStore:n,menuDescriptorStore:i}),vu()}),Ws(()=>{s&&s(),s=null}),(a,r)=>(k(),T(Q,null,[q(dc),q(Vd),q(jd),q(nu),q(su),q(ou),q(pu),q(du)],64))}},bu={class:"sheet-columns"},_u={class:"sheet-left"},wu={class:"sheet-main"},Su=["id"],Eu={class:"panel concept-sheet-panel"},xu=["id"],Iu=["id"],Cu=["id"],ku={__name:"AnalysisSheetShell",props:{counter:{type:Number,required:!0}},setup(e){const t=qe(),n=Ne(()=>t.argPanelWidth?t.argPanelStyle:{flex:"0 0 30%"}),i=Ne(()=>({flex:t.statePanelWidth?`0 0 ${t.statePanelWidth}px`:"0 0 220px",minWidth:"180px",overflowY:"auto"}));return(s,a)=>(k(),T("div",bu,[f("div",_u,[f("div",wu,[f("div",{class:"panel",style:Te(n.value)},[a[2]||(a[2]=f("div",{class:"panel-header"},[f("strong",{class:"column-title"},"ARG (Abstract Reachability Graph)")],-1)),f("div",{id:`arg-graph-${e.counter}`,class:"graph-container",onContextmenu:a[0]||(a[0]=it(()=>{},["prevent"]))},null,40,Su)],4),a[4]||(a[4]=f("div",{class:"divider"},null,-1)),f("div",Eu,[a[3]||(a[3]=f("div",{class:"panel-header"},[f("strong",{class:"column-title"},"Concept graph")],-1)),f("div",{id:`concept-graph-${e.counter}`,class:"graph-container",onContextmenu:a[1]||(a[1]=it(()=>{},["prevent"]))},null,40,xu)])]),f("div",{class:"info-panel",style:Te(g(t).detailsPanelStyle)},[f("div",{id:`info-header-${e.counter}`,class:"info-header",title:"Drag to resize Details"},"Details",8,Iu),f("div",{id:`info-content-${e.counter}`},"Select a node or edge to see details",8,Cu)],4)]),a[6]||(a[6]=f("div",{class:"divider"},null,-1)),f("div",{class:"panel",style:Te(i.value)},[...a[5]||(a[5]=[f("div",{class:"panel-header"},[f("strong",{class:"column-title"},"State/relations")],-1),f("div",{class:"state-controls-placeholder"},null,-1)])],4)]))}};function Au({legacyApp:e,editorStore:t,doc:n=globalThis.document,codeMirror:i=globalThis.CodeMirror}={}){const s=n&&n.getElementById("model-editor");if(!s)return null;if(s.__ivyCodeMirrorEditor)return s.__ivyCodeMirrorEditor;if(!i||typeof i.fromTextArea!="function")throw new Error("CodeMirror is not available");const a=i.fromTextArea(s,{lineNumbers:!0,keyMap:t&&t.keymap?t.keymap:"sublime",tabSize:4,indentUnit:4,lineWrapping:!1,matchBrackets:!0,extraKeys:{"Ctrl-Z":"undo","Ctrl-Y":"redo","Ctrl-Shift-Z":"redo"}});return s.__ivyCodeMirrorEditor=a,a&&typeof a.on=="function"&&e&&a.on("change",()=>{!e.cmEditor||typeof e.cmEditor.getValue!="function"||(e._persistedFileContent=e.cmEditor.getValue(),typeof e._updateEditorLabel=="function"&&e._updateEditorLabel())}),a}function Fi(e,t=0,n=0,i=globalThis.document){const s=i&&i.getElementById("context-menu");s&&(s.style.display=e?"block":"none",e&&(s.style.left=`${Number(t)||0}px`,s.style.top=`${Number(n)||0}px`))}function Du({app:e,pinia:t,doc:n=globalThis.document,win:i=globalThis.window,hFn:s=dl,nextTickFn:a=Et,renderFn:r=Rl}={}){if(!t)throw new Error("createIvyVueBridge requires a Pinia instance");const o=Qa(t),l=wt(t),d=Js(t),c=tr(t),h=ir(t),m=xt(t),E=ns(t),B=sr(t),D=Xs(t),W=is(t),U=mu(t),V=Ya(t),H=ss(t),N=qe(t),Y=ar(t);return{createLegacyApi(){return new Za(l.engine)},getEngine(){return l.engine},showContextMenu(y,P,G){d.show(y,P,G),Fi(!0,y,P,n)},hideContextMenu(){d.hide(),Fi(!1,0,0,n)},updateEditor(y){o.applyLegacySnapshot(y)},initializeEditor(y){return Au({legacyApp:y,editorStore:o,doc:n,codeMirror:i.CodeMirror})},setEditorKeymap(y){o.setKeymap(y)},getEditorKeymap(){return o.keymap},editorKeymapHandled(){return!0},updateReopenLastFileButton(y,P){o.setReopenLastFileButton(y,P)},staticCommandHandlersHandled(){return!0},tabClicksHandled(){return!0},fileInputHandlersHandled(){return!0},layoutResizersHandled(){return!0},globalInteractionsHandled(){return!0},graphContextMenuSuppressionHandled(){return!0},updateDetails(y){c.setDetails(y)},clearDetails(){c.clear()},updateConstraintFacts(y,P){c.setConstraintFacts(y,P)},setCheckTraceAction(y){c.setTraceAction(y)},updateStateRelations(y,P){B.setRows(y,P)},getStateRelationToggles(){return B.toggleSnapshot},setStateRelationToggles(y){B.applyToggleSnapshot(y||{})},buildStateRelationVisibility(){return B.visibilitySnapshot},clearStateRelations(){B.clear()},updateStateLabel(y){B.setStateLabel(y)},updateStatus(y,P=""){E.setStatus(y,P)},setSessionId(y){E.setSessionId(y)},setMode(y){E.setMode(y)},getMode(){return E.mode},setLoadedFile(y,P){E.setLoadedFile(y,P)},showLoading(y){E.showLoading(y)},hideLoading(){E.hideLoading()},showToast(y,P,G){return Y.show(y,P,G||{})},setSaveAsNoticeVisible(y){E.setSaveAsNoticeVisible(y)},updateMenuRegion(y,P,G){D.setRegion(y,P,G)},closeDropdownMenus(){D.closeAll(),m.closeAll()},flashMenuItem(y,P){m.flashItem(y,P)},upsertSheetTab(y){W.upsertTab(y)},createAnalysisSheetShell({id:y,counter:P}){if(!y)return null;const G=n&&n.getElementById("sheet-area");if(!G)return null;let z=n.getElementById(y);z||(z=n.createElement("div"),z.id=y,z.className="sheet-content",z.__ivyVueRenderedSheet=!0,G.appendChild(z));const le=s(ku,{counter:P});return le.appContext=e&&e._context,r(le,z),z},removeRenderedSheet(y){const P=n&&n.getElementById(y);return P&&P.__ivyVueRenderedSheet?(r(null,P),!0):!1},activateSheetTab(y){W.activateTab(y)},removeSheetTab(y){W.removeTab(y),H.removeSheet(y)},resetSheetTabs(){W.resetTabs(),H.reset()},getSheetTabLabel(y){return W.labelFor(y)},setActiveGraphSheet(y){U.setActiveSheet(y)},updateGraphSnapshot(y,P,G){U.applyGraphSnapshot(y,P,G)},selectArgNode(y,P){U.selectArgNode(P,y)},showDialog(y){return h.open(y)},editorSaveSheenHandled(){return!0},updateRecentFiles(y,P){V.setItems(y,P)},upsertEventTraceSheet(y){H.upsertSheet(y)},setEventTraceExpanded(y,P,G){H.setExpanded(y,P,G)},isEventTraceExpanded(y,P){return H.isExpanded(y,P)},selectEventTraceRow(y,P){H.selectEvent(y,P)},updateEventPatterns(y,P){H.setPatterns(y,P)},setSelectedEventPatternIndex(y,P){H.setSelectedPatternIndex(y,P)},getSelectedEventPattern(y){return H.selectedPattern(y)},getSelectedEventPatternIndex(y){return H.selectedPatternIndex(y)},setTutorialVisible(y){N.setTutorialVisible(y)},isTutorialVisible(){return N.tutorialVisible},flashTutorialButton(y){N.flashTutorialButton(y)},setArgPanelWidth(y){N.setArgPanelWidth(y)},setStatePanelWidth(y){N.setStatePanelWidth(y)},setDetailsHeight(y){N.setDetailsHeight(y)},setEditorWidth(y){N.setEditorWidth(y)},setTutorialHeight(y){N.setTutorialHeight(y)},tutorialUrlBarHandled(){return!0},afterLayoutSettled(y){typeof y=="function"&&a(()=>{const P=typeof i.requestAnimationFrame=="function"?i.requestAnimationFrame.bind(i):G=>i.setTimeout(G,0);P(()=>P(y))})}}}function Tu({pinia:e,win:t=globalThis.window}={}){if(!e||!t)return null;const n=wt(e);if(t.__IVY_ENGINE__){const i=t.__IVY_ENGINE_KIND__||t.__IVY_ENGINE__.kind||"custom";return n.setEngine(i,t.__IVY_ENGINE__)}return t.__IVY_ENGINE_KIND__==="wanix"?n.useWanix():n.engine}function Lu(e={}){const t=e.win||globalThis.window;Tu({pinia:e.pinia,win:t});const n=Du(e);return t.__ivyVueBridge=n,n}const Pu=`/**
 * IvyApp - Main application for the Ivy Interactive Verification web UI.
 *
 * Orchestrates the API, graph views, and UI controls. Handles all
 * user interactions including ARG/concept graph clicks, context menus,
 * file loading, mode selection, and verification checks.
 */

'use strict';

class IvyApp {
    constructor() {
        this.api = this.createApi();
        this.controls = new IvyControls(this.api);
        this.argGraph = null;
        this.conceptGraph = null;
        this.sheets = {};
        this.activeSheetId = 'sheet-1';
        this.selectedArgNode = null;
        // Edge visibility state, matching Python's edge_display_checkboxes.
        // Keys: edgeName, values: {all_to_all: bool, edge_unknown: bool, none_to_none: bool, transitive: bool}
        // Default: all false (edges hidden until checkbox is checked).
        this._edgeVisibility = {};
        // Node label visibility state, matching Python's node_label_display_checkboxes.
        // Keys: labelName, values: {node_necessarily: bool, node_maybe: bool, node_necessarily_not: bool}
        // Maps to checkbox columns: + → node_necessarily, ? → node_maybe, - → node_necessarily_not
        this._labelVisibility = {};
        // Concept data from last server response (for node/label sort info)
        this._lastConceptData = null;
        // File System Access API handle for in-place saves (Ctrl+S).
        this._fileHandle = null;
        this._lastClosedFileHandle = null;
        this._lastClosedSessionId = '';
        this._lastClosedFileName = '';
        // Content as last written to disk; used to detect unsaved changes.
        this._savedFileContent = null;
        this._saveInProgress = false;
        this._saveProgressSheen = null;
        this.currentBound = 10;
    }

    createApi() {
        var bridge = window.__ivyVueBridge;
        if (bridge && typeof bridge.createLegacyApi === 'function') {
            try {
                return bridge.createLegacyApi();
            } catch (e) {
                console.warn('Vue engine bridge unavailable, falling back to IvyAPI:', e);
            }
        }
        return new IvyAPI();
    }

    /**
     * Initialize the application: create session, build graphs, wire events.
     */
    async init() {
        var self = this;
        this.controls.setStatus('Initializing...');

        // Check for a saved session BEFORE creating a new server session.
        // This prevents the URL session ID from incrementing on every reload.
        var savedState = IvyPersist.load();

        // Always need a server session for API calls.
        try {
            await this.api.createSession();
        } catch (e) {
            this.controls.setStatus('Failed to create session: ' + e.message, 'error');
            console.error('Session creation failed:', e);
        }

        // If restoring, keep the saved session's URL hash.
        // If fresh, set the new session ID in the URL.
        if (!savedState || !savedState.fileContent) {
            IvyPersist.setSessionIdInURL(this.api.sessionId);
        }

        // Show the persisted session ID (from URL hash), not the server session ID.
        // These can differ because the server ID increments on restart while
        // the persisted ID is stable across reloads.
        this.updateSessionDisplay(IvyPersist.getSessionIdFromURL() || this.api.sessionId);

        // Create Cytoscape graph instances
        this.argGraph = new IvyGraph('arg-graph', ARG_STYLE);
        this.conceptGraph = new IvyGraph('concept-graph', CONCEPT_STYLE);
        this.registerSheet('sheet-1', this.argGraph, this.conceptGraph);

        // Health check: verify graphs initialized correctly.
        this.argGraph.healthCheck();
        this.conceptGraph.healthCheck();

        // Hook concept graph updates to auto-apply edge visibility.
        // Matches Python: edges are hidden by default, shown only when
        // the corresponding checkbox (+/?/-) in the state panel is checked.
        // Wire up all event handlers
        this.setupEventHandlers();
        await this.loadMenuDescriptors();
        this.setupTabs();
        if (!(window.__ivyVueBridge &&
            typeof window.__ivyVueBridge.layoutResizersHandled === 'function' &&
            window.__ivyVueBridge.layoutResizersHandled())) {
            this.setupResizer();
            this.setupResizer2();
            this.setupResizer3();
            this.setupResizerH();
            this.setupDetailsResizer();
        }
        this.setupTutorialUrlBar();
        this.setupKeyboardShortcuts();

        // Connect to SSE for real-time updates
        if (this.api.sessionId) {
            this.api.connectEvents(this.handleEvent.bind(this));
            this.api.onConnectionLost = function () {
                self.controls.setStatus('Server connection lost', 'error');
                self._showToast('Connection to server lost. Check that the server is running and reload the page.', 'error');
            };
        }

        // Initialize CodeMirror on the model editor textarea.
        // Must happen BEFORE restore so setEditorContent() can call cmEditor.setValue().
        var modelEditor = document.getElementById('model-editor');
        if (window.__ivyVueBridge && typeof window.__ivyVueBridge.initializeEditor === 'function') {
            this.cmEditor = window.__ivyVueBridge.initializeEditor(this);
        } else if (modelEditor) {
            this.cmEditor = CodeMirror.fromTextArea(modelEditor, {
                lineNumbers: true,
                keyMap: this.getEditorKeymap(),
                tabSize: 4,
                indentUnit: 4,
                lineWrapping: false,
                matchBrackets: true,
                extraKeys: {
                    'Ctrl-Z': 'undo',
                    'Ctrl-Y': 'redo',
                    'Ctrl-Shift-Z': 'redo',
                }
            });
            // Sync edits back to persisted content and update dirty marker.
            this.cmEditor.on('change', function () {
                if (!self.cmEditor || typeof self.cmEditor.getValue !== 'function') return;
                self._persistedFileContent = self.cmEditor.getValue();
                self._updateEditorLabel();
            });
            // Keymap radio button switching fallback for non-Vue test harnesses.
            if (!(window.__ivyVueBridge &&
                typeof window.__ivyVueBridge.editorKeymapHandled === 'function' &&
                window.__ivyVueBridge.editorKeymapHandled())) {
                var radios = document.querySelectorAll('input[name="keymap"]');
                for (var i = 0; i < radios.length; i++) {
                    radios[i].addEventListener('change', function () {
                        self.setEditorKeymap(this.value);
                    });
                }
            }
        }

        // Restore saved session if available (survives page reload).
        if (savedState && savedState.fileContent) {
            console.log('IvyPersist: restoring session', savedState.sessionId, savedState.fileName);
            var restored = await IvyPersist.restore(this, savedState);
            if (restored) {
                // Keep the URL hash from the saved session (don't overwrite)
                IvyPersist.setFileName(savedState.fileName, savedState.filePath);
                this.controls.setStatus('Restored: ' + (savedState.fileName || 'session'), 'success');
            } else {
                this.controls.setStatus('Ready');
            }
        } else {
            this.controls.setStatus('Ready');
        }

        // Auto-save: on beforeunload (catches reload, tab close, navigation)
        // and after any successful operation (debounced).
        window.addEventListener('beforeunload', function () {
            IvyPersist.save(self);
        });

        // Hook into setStatus: auto-save whenever a 'success' status is set.
        var origSetStatus = this.controls.setStatus.bind(this.controls);
        this.controls.setStatus = function (msg, level) {
            origSetStatus(msg, level);
            if (level === 'success') {
                IvyPersist.save(self);
            }
        };
    }

    // --- Editor helpers (CodeMirror) ---

    updateSessionDisplay(sessionId) {
        var displaySessionId = sessionId || '';
        if (window.__ivyVueBridge && typeof window.__ivyVueBridge.setSessionId === 'function') {
            window.__ivyVueBridge.setSessionId(displaySessionId);
            return;
        }
        var sessionEl = document.getElementById('session-id');
        if (sessionEl) {
            sessionEl.textContent = displaySessionId ? 'Session: ' + displaySessionId : '';
        }
    }

    _setVueLayoutSize(method, value) {
        var bridge = window.__ivyVueBridge;
        if (bridge && typeof bridge[method] === 'function') {
            bridge[method](value);
            return true;
        }
        return false;
    }

    setEditorContent(content) {
        this._persistedFileContent = content;
        this._savedFileContent = content;
        if (this.cmEditor) {
            this.cmEditor.setValue(content);
        }
        this._updateEditorLabel();
    }

    _updateEditorLabel() {
        var name = this._persistedFilePath || this._persistedFileName || '';
        var current = (this.cmEditor && typeof this.cmEditor.getValue === 'function') ? this.cmEditor.getValue() : (this._persistedFileContent || '');
        var saved = this._savedFileContent || '';
        var dirty = current !== saved;
        var labelText;
        if (!name) {
            labelText = '(unsaved file)' + (this._saveInProgress ? ' [saving...]' : '');
            if (dirty && !this._saveInProgress) {
                labelText = '** ' + labelText;
            }
        } else if (this._saveInProgress) {
            labelText = name + ' [saving...]';
        } else if (dirty) {
            labelText = '** ' + name;
        } else {
            labelText = name + ' [saved]';
        }
        if (window.__ivyVueBridge && typeof window.__ivyVueBridge.updateEditor === 'function') {
            window.__ivyVueBridge.updateEditor({
                path: name,
                content: current,
                savedContent: saved,
                saveInProgress: this._saveInProgress,
            });
            this._updateReopenLastFileButton();
            return;
        }
        var editorLabel = document.getElementById('model-editor-label');
        if (!editorLabel) return;
        editorLabel.textContent = labelText;
        editorLabel.title = 'Editing: ' + labelText;
        this._updateReopenLastFileButton();
    }

    _refreshEditorLayout() {
        var self = this;
        var refresh = function () {
            if (self.cmEditor && typeof self.cmEditor.refresh === 'function') {
                self.cmEditor.refresh();
            }
        };

        refresh();
        if (window.requestAnimationFrame) {
            window.requestAnimationFrame(refresh);
        }
        setTimeout(refresh, 0);
    }

    _editorContent() {
        return this.cmEditor ? this.cmEditor.getValue() : (this._persistedFileContent || '');
    }

    _editorDirty() {
        return this._editorContent() !== (this._savedFileContent || '');
    }

    _showSaveProgress(message) {
        this._hideSaveProgress();
        this._saveInProgress = true;
        this._updateEditorLabel();
        this._saveProgressSheen = this._showSaveEditorSheen();
        return true;
    }

    _editorTextElement() {
        if (this.cmEditor && typeof this.cmEditor.getWrapperElement === 'function') {
            var wrapper = this.cmEditor.getWrapperElement();
            if (wrapper) return wrapper;
        }
        var editorPanel = document.getElementById('editor-panel');
        if (editorPanel) {
            return editorPanel.querySelector('.CodeMirror') || editorPanel.querySelector('#model-editor');
        }
        return document.getElementById('model-editor');
    }

    _showSaveEditorSheen() {
        if (window.__ivyVueBridge && typeof window.__ivyVueBridge.editorSaveSheenHandled === 'function' && window.__ivyVueBridge.editorSaveSheenHandled()) {
            return { vueHandled: true };
        }
        var target = this._editorTextElement();
        if (!target || typeof target.getBoundingClientRect !== 'function') return null;
        var rect = target.getBoundingClientRect();
        if (!rect || rect.width <= 0 || rect.height <= 0) return null;
        var sheen = document.createElement('div');
        sheen.className = 'ivy-save-editor-sheen';
        sheen.setAttribute('aria-hidden', 'true');
        sheen.style.cssText = 'position:fixed;z-index:9999;pointer-events:none;' +
            'background:rgba(128,128,128,0.2);';
        sheen.style.left = Math.round(rect.left) + 'px';
        sheen.style.top = Math.round(rect.top) + 'px';
        sheen.style.width = Math.round(rect.width) + 'px';
        sheen.style.height = Math.round(rect.height) + 'px';
        document.body.appendChild(sheen);
        return sheen;
    }

    _hideSaveProgress() {
        if (this._saveProgressSheen && this._saveProgressSheen.parentNode) {
            this._saveProgressSheen.parentNode.removeChild(this._saveProgressSheen);
        }
        this._saveProgressSheen = null;
        if (this._saveInProgress) {
            this._saveInProgress = false;
            this._updateEditorLabel();
        }
    }

    _rememberLastOpenFile() {
        var name = this._persistedFileName || (this._fileHandle && this._fileHandle.name) || '';
        if (!name) return;
        this._lastClosedFileHandle = this._fileHandle;
        this._lastClosedSessionId = IvyPersist.getSessionIdFromURL() || (this.api && this.api.sessionId) || '';
        this._lastClosedFileName = name || 'file';
    }

    _updateReopenLastFileButton() {
        var btn = document.getElementById('file-reopen-last');
        var noCurrentFile = !this._fileHandle && !this._persistedFileName;
        var visible = !!(noCurrentFile && (this._lastClosedFileHandle || this._lastClosedSessionId));
        var label = 'Re-open last file ' + (this._lastClosedFileName || 'file');
        if (window.__ivyVueBridge && typeof window.__ivyVueBridge.updateReopenLastFileButton === 'function') {
            window.__ivyVueBridge.updateReopenLastFileButton(visible, label);
            return;
        }
        if (!btn) return;
        if (visible) {
            btn.textContent = label;
            btn.style.display = '';
        } else {
            btn.style.display = 'none';
        }
    }

    async reopenLastFile() {
        if (!this._lastClosedFileHandle && !this._lastClosedSessionId) return;
        try {
            if (this._lastClosedFileHandle) {
                var file = await this._lastClosedFileHandle.getFile();
                this._fileHandle = this._lastClosedFileHandle;
                await this.loadFile(file);
            } else {
                await this.loadRecentSession(this._lastClosedSessionId);
            }
            IvyPersist.setFileName(
                this._persistedFileName || this._lastClosedFileName,
                this._persistedFilePath || this._persistedFileName || this._lastClosedFileName
            );
        } catch (e) {
            this.controls.setStatus('Re-open failed: ' + e.message, 'error');
        }
    }

    async _ensureFileHandleWritable() {
        if (!this._fileHandle) return false;
        if (!this._fileHandle.queryPermission || !this._fileHandle.requestPermission) {
            return true;
        }
        var opts = { mode: 'readwrite' };
        var perm = await this._fileHandle.queryPermission(opts);
        if (perm === 'granted') return true;
        perm = await this._fileHandle.requestPermission(opts);
        return perm === 'granted';
    }

    async _readFileHandleContent() {
        if (!this._fileHandle) return null;
        var file = await this._fileHandle.getFile();
        return await file.text();
    }

    async _restoreFileHandleForCurrentFile() {
        if (this._fileHandle) return true;
        if (!this._persistedFileName) return false;
        var state = {
            sessionId: IvyPersist.getSessionIdFromURL() || (this.api && this.api.sessionId) || '',
            fileName: this._persistedFileName,
            filePath: this._persistedFilePath || this._persistedFileName,
        };
        var handle = await IvyPersist.loadFileHandle(state);
        if (!handle) return false;
        this._fileHandle = handle;
        await IvyPersist.saveFileHandle(this);
        return true;
    }

    async _confirmNoExternalChangeBeforeSave(content) {
        if (!this._fileHandle) return 'ok';
        var diskContent = await this._readFileHandleContent();
        var lastSaved = this._savedFileContent || '';
        if (diskContent === lastSaved || diskContent === content) {
            return 'ok';
        }
        var choice = await this.showExternalChangeDialog();
        if (choice === 'overwrite') {
            return 'overwrite';
        }
        if (choice === 'reload') {
            this.setEditorContent(diskContent);
            this._persistedFileContent = diskContent;
            this._savedFileContent = diskContent;
            IvyPersist.save(this);
            this.controls.setStatus('Reverted to on-disk version: ' + (this._persistedFileName || 'model'), 'success');
        } else if (choice === 'merge') {
            var merged = this._mergeDiskVersionIntoEditBuffer(lastSaved, content, diskContent);
            this._persistedFileContent = merged;
            this._savedFileContent = diskContent;
            if (this.cmEditor) {
                this.cmEditor.setValue(merged);
            }
            this._updateEditorLabel();
            IvyPersist.save(this);
            this.controls.setStatus('Merged disk changes into editor buffer; resolve conflict markers before saving', 'warning');
        } else {
            this.controls.setStatus('Save cancelled: file changed on disk', 'warning');
        }
        return 'skip';
    }

    _mergeDiskVersionIntoEditBuffer(baseContent, editorContent, diskContent) {
        if (editorContent === baseContent) return diskContent;
        if (diskContent === baseContent) return editorContent;
        return [
            '<<<<<<< EDIT BUFFER',
            editorContent.replace(/\\s*$/, ''),
            '||||||| LAST SAVED',
            baseContent.replace(/\\s*$/, ''),
            '=======',
            diskContent.replace(/\\s*$/, ''),
            '>>>>>>> ON DISK',
            ''
        ].join('\\n');
    }

    showExternalChangeDialog() {
        if (window.__ivyVueBridge && typeof window.__ivyVueBridge.showDialog === 'function') {
            return window.__ivyVueBridge.showDialog({
                type: 'buttons',
                title: 'File changed on disk',
                message: 'The file "' + (this._persistedFileName || 'model') + '" has changed outside IvyWeb. Choose how to handle the current editor buffer.',
                escapeValue: 'do-nothing',
                buttons: [
                    { label: 'Do nothing', value: 'do-nothing' },
                    { label: 'Revert to on-disk version', value: 'reload' },
                    { label: 'Merge disk version into edit buffer', value: 'merge' },
                    { label: 'Overwrite on-disk with edited buffer', value: 'overwrite', danger: true },
                ],
            });
        }
        var self = this;
        return new Promise(function (resolve) {
            var overlay = document.getElementById('external-change-dialog-overlay');
            var msg = document.getElementById('external-change-dialog-message');
            var doNothing = document.getElementById('external-change-do-nothing');
            var reload = document.getElementById('external-change-reload');
            var merge = document.getElementById('external-change-merge');
            var overwrite = document.getElementById('external-change-overwrite');
            if (!overlay || !msg || !doNothing || !reload || !merge || !overwrite) {
                resolve('do-nothing');
                return;
            }
            msg.textContent = 'The file "' + (self._persistedFileName || 'model') + '" has changed outside IvyWeb. Choose how to handle the current editor buffer.';
            overlay.style.display = 'flex';
            doNothing.focus();

            var done = function (choice) {
                overlay.style.display = 'none';
                doNothing.removeEventListener('click', onDoNothing);
                reload.removeEventListener('click', onReload);
                merge.removeEventListener('click', onMerge);
                overwrite.removeEventListener('click', onOverwrite);
                document.removeEventListener('keydown', onKeyDown);
                resolve(choice);
            };
            var onDoNothing = function () { done('do-nothing'); };
            var onReload = function () { done('reload'); };
            var onMerge = function () { done('merge'); };
            var onOverwrite = function () { done('overwrite'); };
            var onKeyDown = function (e) {
                if (e.key === 'Escape' || e.key === 'Enter') {
                    e.preventDefault();
                    done('do-nothing');
                }
            };
            doNothing.addEventListener('click', onDoNothing);
            reload.addEventListener('click', onReload);
            merge.addEventListener('click', onMerge);
            overwrite.addEventListener('click', onOverwrite);
            document.addEventListener('keydown', onKeyDown);
        });
    }

    showDirtyCloseDialog() {
        if (window.__ivyVueBridge && typeof window.__ivyVueBridge.showDialog === 'function') {
            return window.__ivyVueBridge.showDialog({
                type: 'buttons',
                title: 'File has changed',
                message: 'File "' + (this._persistedFileName || 'model') + '" has changed. Save before closing?',
                escapeValue: 'cancel',
                buttons: [
                    { label: 'Cancel the close', value: 'cancel' },
                    { label: 'Save', value: 'save' },
                    { label: 'Discard Edits', value: 'discard', danger: true },
                ],
            });
        }
        var self = this;
        return new Promise(function (resolve) {
            var overlay = document.getElementById('dirty-close-dialog-overlay');
            var msg = document.getElementById('dirty-close-dialog-message');
            var cancel = document.getElementById('dirty-close-cancel');
            var save = document.getElementById('dirty-close-save');
            var discard = document.getElementById('dirty-close-discard');
            if (!overlay || !msg || !cancel || !save || !discard) {
                resolve('cancel');
                return;
            }
            msg.textContent = 'File "' + (self._persistedFileName || 'model') + '" has changed. Save before closing?';
            overlay.style.display = 'flex';
            cancel.focus();

            var done = function (choice) {
                overlay.style.display = 'none';
                cancel.removeEventListener('click', onCancel);
                save.removeEventListener('click', onSave);
                discard.removeEventListener('click', onDiscard);
                document.removeEventListener('keydown', onKeyDown);
                resolve(choice);
            };
            var onCancel = function () { done('cancel'); };
            var onSave = function () { done('save'); };
            var onDiscard = function () { done('discard'); };
            var onKeyDown = function (e) {
                if (e.key === 'Enter' || e.key === 'Escape') {
                    e.preventDefault();
                    done('cancel');
                }
            };
            cancel.addEventListener('click', onCancel);
            save.addEventListener('click', onSave);
            discard.addEventListener('click', onDiscard);
            document.addEventListener('keydown', onKeyDown);
        });
    }

    showSaveAsExplanationNotice() {
        if (window.__ivyVueBridge && typeof window.__ivyVueBridge.setSaveAsNoticeVisible === 'function') {
            window.__ivyVueBridge.setSaveAsNoticeVisible(true);
            return;
        }
        var notice = document.getElementById('save-as-explain-notice');
        if (notice) {
            notice.style.display = 'block';
        }
    }

    hideSaveAsExplanationNotice() {
        if (window.__ivyVueBridge && typeof window.__ivyVueBridge.setSaveAsNoticeVisible === 'function') {
            window.__ivyVueBridge.setSaveAsNoticeVisible(false);
            return;
        }
        var notice = document.getElementById('save-as-explain-notice');
        if (notice) {
            notice.style.display = 'none';
        }
    }

    scrollEditorToLine(lineno) {
        if (this.cmEditor) {
            var line = lineno - 1;
            if (this._highlightedEditorLineHandle != null) {
                this.cmEditor.removeLineClass(this._highlightedEditorLineHandle, 'background', 'ivy-source-highlight');
            }
            this.cmEditor.setCursor(line, 0);
            this.cmEditor.setSelection(
                {line: line, ch: 0},
                {line: line, ch: this.cmEditor.getLine(line).length}
            );
            this._highlightedEditorLineHandle = this.cmEditor.addLineClass(line, 'background', 'ivy-source-highlight');
            this._highlightedEditorLine = lineno;
            this.cmEditor.scrollIntoView({line: line, ch: 0}, 50);
            this.cmEditor.focus();
        }
    }

    currentSheet() {
        return this.sheets ? this.sheets[this.activeSheetId] : null;
    }

    registerSheet(sheetId, argGraph, conceptGraph) {
        this.installConceptGraphVisibilityHook(conceptGraph);
        this.installGraphStoreHook(sheetId, 'arg', argGraph);
        this.installGraphStoreHook(sheetId, 'concept', conceptGraph);
        this.sheets[sheetId] = {
            id: sheetId,
            type: 'analysis',
            argGraph: argGraph,
            conceptGraph: conceptGraph,
            selectedArgNode: null,
            visualOnly: false,
        };
    }

    installGraphStoreHook(sheetId, kind, graph) {
        if (!graph || graph._ivyGraphStoreHooked) return;
        var origUpdate = graph.update.bind(graph);
        graph.update = function(elements, positions) {
            var result = origUpdate(elements, positions);
            if (window.__ivyVueBridge && typeof window.__ivyVueBridge.updateGraphSnapshot === 'function') {
                window.__ivyVueBridge.updateGraphSnapshot(sheetId, kind, {
                    elements: elements || [],
                    positions: positions || null,
                });
            }
            return result;
        };
        graph._ivyGraphStoreHooked = true;
    }

    isVisualOnlySheet(sheetId) {
        var sheet = this.sheets && this.sheets[sheetId];
        return !!(sheet && sheet.visualOnly);
    }

    setVisualOnlySheet(sheetId, visualOnly) {
        var sheet = this.sheets && this.sheets[sheetId];
        if (sheet) sheet.visualOnly = !!visualOnly;
    }

    visualOnlyMessage(kind) {
        if (kind === 'events') {
            return 'Restored event trace is visual-only; reload or rerun analysis before event backend actions';
        }
        return 'Restored analysis state is visual-only; reload or rerun analysis before graph actions';
    }

    installConceptGraphVisibilityHook(conceptGraph) {
        if (!conceptGraph || conceptGraph._ivyVisibilityHooked) return;
        var self = this;
        var origUpdate = conceptGraph.update.bind(conceptGraph);
        conceptGraph.update = function(elements, positions) {
            origUpdate(elements, positions);
            if (conceptGraph.cy && typeof conceptGraph.cy.edges === 'function') {
                self._applyEdgeVisibility(conceptGraph);
            }
        };
        conceptGraph._ivyVisibilityHooked = true;
    }

    attachGraphEventHandlers(argGraph, conceptGraph, sheetId) {
        var self = this;
        sheetId = sheetId || this.activeSheetId || 'sheet-1';
        var vueSuppressesNativeGraphMenu = window.__ivyVueBridge &&
            typeof window.__ivyVueBridge.graphContextMenuSuppressionHandled === 'function' &&
            window.__ivyVueBridge.graphContextMenuSuppressionHandled();
        if (!vueSuppressesNativeGraphMenu) {
            if (argGraph && argGraph.containerId) {
                var argEl = document.getElementById(argGraph.containerId);
                if (argEl && !argEl._ivyContextSuppressed) {
                    argEl.addEventListener('contextmenu', function (e) { e.preventDefault(); });
                    argEl._ivyContextSuppressed = true;
                }
            }
            if (conceptGraph && conceptGraph.containerId) {
                var conceptEl = document.getElementById(conceptGraph.containerId);
                if (conceptEl && !conceptEl._ivyContextSuppressed) {
                    conceptEl.addEventListener('contextmenu', function (e) { e.preventDefault(); });
                    conceptEl._ivyContextSuppressed = true;
                }
            }
        }

        argGraph.onNodeClick(function (nodeData) {
            self.onArgNodeClick(nodeData, sheetId);
        });
        argGraph.onNodeRightClick(function (nodeData, pos) {
            self.onArgNodeRightClick(nodeData, pos, sheetId);
        });
        argGraph.onEdgeClick(function (edgeData) {
            self.controls.showInfo(edgeData.short_info, edgeData.long_info);
        });
        argGraph.onEdgeRightClick(function (edgeData, pos) {
            self.onArgEdgeRightClick(edgeData, pos, sheetId);
        });
        argGraph.onBackgroundClick(function () {
            self.controls.clearInfo();
            self.controls.hideContextMenu();
        });

        conceptGraph.onNodeRightClick(function (nodeData, pos) {
            self.onConceptNodeRightClick(nodeData, pos);
        });
        conceptGraph.onNodeClick(function (nodeData, evt) {
            try {
                var node = evt.target;
                var name = nodeData.obj || nodeData.id;
                if (node.hasClass('selected_node')) {
                    node.removeClass('selected_node');
                    node.unselect();
                    self.controls.setStatus('Deselected: ' + name);
                    self.controls.clearInfo();
                } else {
                    node.addClass('selected_node');
                    self.selectedConceptNode = name;
                    self.controls.setStatus('Selected: ' + name);
                    self.controls.showInfo(nodeData.short_info, nodeData.long_info);
                }
            } catch (e) {
                console.error('concept node click error:', e);
            }
        });
        conceptGraph.onEdgeClick(function (edgeData, evt) {
            var edge = evt.target;
            var name = edgeData.obj || edgeData.label || edgeData.id;
            if (edge.hasClass('selected_edge')) {
                edge.removeClass('selected_edge');
                edge.unselect();
                edge.removeStyle('line-color target-arrow-color source-arrow-color width');
                self.controls.setStatus('Deselected: ' + name);
                self.controls.clearInfo();
            } else {
                edge.addClass('selected_edge');
                self.controls.setStatus('Selected: ' + name);
                self.controls.showInfo(edgeData.short_info, edgeData.long_info);
            }
        });
        conceptGraph.onEdgeRightClick(function (edgeData, pos) {
            self.onConceptEdgeRightClick(edgeData, pos);
        });
        conceptGraph.onBackgroundClick(function () {
            self.controls.clearInfo();
            self.controls.hideContextMenu();
        });
    }

    async chooseAndLoadModelFile() {
        var fileInput = document.getElementById('file-input');
        if (window.showOpenFilePicker) {
            try {
                var handles = await window.showOpenFilePicker({
                    types: [{ description: 'Ivy files', accept: { 'text/plain': ['.ivy'] } }],
                    multiple: false,
                });
                var handle = handles[0];
                var file = await handle.getFile();
                this._fileHandle = handle;
                await this.loadFile(file);
            } catch (ex) {
                if (ex.name !== 'AbortError') {
                    this.controls.setStatus('Load failed: ' + ex.message, 'error');
                }
            }
        } else if (fileInput) {
            fileInput.click();
        }
    }

    async chooseAndLoadEventTraceFile() {
        var eventFileInput = document.getElementById('event-file-input');
        if (window.showOpenFilePicker) {
            try {
                var handles = await window.showOpenFilePicker({
                    types: [{ description: 'Ivy event traces', accept: { 'text/plain': ['.iev', '.txt'] } }],
                    multiple: false,
                });
                var file = await handles[0].getFile();
                await this.loadEventTraceFile(file);
            } catch (ex) {
                if (ex.name !== 'AbortError') {
                    this.controls.setStatus('Event trace load failed: ' + ex.message, 'error');
                }
            }
        } else if (eventFileInput) {
            eventFileInput.click();
        }
    }

    chooseAndLoadAnalysisStateFile() {
        var analysisStateFileInput = document.getElementById('analysis-state-file-input');
        if (analysisStateFileInput) {
            analysisStateFileInput.click();
        }
    }

    /**
     * Set up all DOM and graph event handlers.
     */
    setupEventHandlers() {
        var self = this;

        // --- File Menu ---
        var fileInput = document.getElementById('file-input');
        var eventFileInput = document.getElementById('event-file-input');
        var analysisStateFileInput = document.getElementById('analysis-state-file-input');
        var vueHandlesStaticCommands = window.__ivyVueBridge &&
            typeof window.__ivyVueBridge.staticCommandHandlersHandled === 'function' &&
            window.__ivyVueBridge.staticCommandHandlersHandled();
        var vueHandlesFileInputs = window.__ivyVueBridge &&
            typeof window.__ivyVueBridge.fileInputHandlersHandled === 'function' &&
            window.__ivyVueBridge.fileInputHandlersHandled();

        if (!vueHandlesFileInputs && fileInput) fileInput.addEventListener('change', function () {
            if (fileInput.files && fileInput.files.length > 0) {
                self.loadFile(fileInput.files[0]);
                fileInput.value = ''; // reset for re-selection of same file
            }
        });

        if (!vueHandlesFileInputs && eventFileInput) {
            eventFileInput.addEventListener('change', function () {
                if (eventFileInput.files && eventFileInput.files.length > 0) {
                    self.loadEventTraceFile(eventFileInput.files[0]);
                    eventFileInput.value = '';
                }
            });
        }

        if (!vueHandlesFileInputs && analysisStateFileInput) {
            analysisStateFileInput.addEventListener('change', async function () {
                if (analysisStateFileInput.files && analysisStateFileInput.files.length > 0) {
                    try {
                        await self.loadAnalysisStateFile(analysisStateFileInput.files[0]);
                    } catch (ex) {
                        self.controls.setStatus('Load analysis state failed: ' + ex.message, 'error');
                    }
                    analysisStateFileInput.value = '';
                }
            });
        }

        if (!vueHandlesStaticCommands) {
            // File > Load...
            // Use showOpenFilePicker when available so we get a writable FileSystemFileHandle,
            // enabling Ctrl+S to save directly without re-prompting. Fall back to <input> otherwise.
            var loadModel = document.getElementById('file-load');
            if (loadModel) loadModel.addEventListener('click', function (e) {
                e.preventDefault();
                self.flashAndClose(this, async function () {
                    await self.chooseAndLoadModelFile();
                });
            });

            var eventTraceOpen = document.getElementById('file-open-event-trace');
            if (eventTraceOpen && eventFileInput) {
                eventTraceOpen.addEventListener('click', function (e) {
                    e.preventDefault();
                    self.flashAndClose(this, async function () {
                        await self.chooseAndLoadEventTraceFile();
                    });
                });
            }

            // File > Save as... (uses File System Access API to write to a chosen path)
            var saveAs = document.getElementById('file-save-as');
            if (saveAs) saveAs.addEventListener('click', function (e) {
                e.preventDefault();
                self.flashAndClose(this, function () { self.saveAs(); });
            });

            // File > Download current model (browser download)
            var download = document.getElementById('file-download');
            if (download) download.addEventListener('click', function (e) {
                e.preventDefault();
                self.flashAndClose(this, function () { self.downloadModel(); });
            });

            var saveAnalysis = document.getElementById('file-save-analysis-state');
            if (saveAnalysis) {
                saveAnalysis.addEventListener('click', function (e) {
                    e.preventDefault();
                    self.flashAndClose(this, function () { self.saveAnalysisState(); });
                });
            }
            var loadAnalysis = document.getElementById('file-load-analysis-state');
            if (loadAnalysis && analysisStateFileInput) {
                loadAnalysis.addEventListener('click', function (e) {
                    e.preventDefault();
                    self.flashAndClose(this, function () { self.chooseAndLoadAnalysisStateFile(); });
                });
            }

            // File > New Model
            var newModel = document.getElementById('file-new');
            if (newModel) newModel.addEventListener('click', function (e) {
                e.preventDefault();
                self.flashAndClose(this, function () { self.newModel(); });
            });

            var closeCurrent = document.getElementById('file-close-current');
            if (closeCurrent) {
                closeCurrent.addEventListener('click', function (e) {
                    e.preventDefault();
                    self.closeCurrentFile();
                });
            }
            var reopenLast = document.getElementById('file-reopen-last');
            if (reopenLast) {
                reopenLast.addEventListener('click', function (e) {
                    e.preventDefault();
                    self.reopenLastFile();
                });
            }

            // File > Save Invariant
            var saveInvariant = document.getElementById('file-save-invariant');
            if (saveInvariant) saveInvariant.addEventListener('click', function (e) {
                e.preventDefault();
                self.flashAndClose(this, function () { self.saveInvariant(); });
            });

            // --- Check ---
            document.getElementById('btn-check').addEventListener('click', function () {
                self.runCheck();
            });

            // --- Show reachable states ---
            document.getElementById('btn-show-reachable').addEventListener('click', function () {
                self.showReachableStates();
            });

            // --- Undo ---
            document.getElementById('btn-undo').addEventListener('click', function () {
                self.doUndo();
            });

            // --- Reset Domain ---
            document.getElementById('btn-reset-domain').addEventListener('click', function () {
                self.resetDomain();
            });

            // --- Diagram Domain ---
            document.getElementById('btn-diagram-domain').addEventListener('click', function () {
                self.diagramDomain();
            });

            // --- Toggle Tutorial ---
            document.getElementById('btn-toggle-tutorial').addEventListener('click', function () {
                self.toggleTutorial();
            });

            // --- Dropdown Menus (panel header) ---
            this.setupDropdownMenus();

            // --- ARG Panel Menu Items (File, Invariant) ---
            this.bindMenuAction('arg-save-abs', function () { self.saveAbstraction(); });
            this.bindMenuAction('arg-check-induction', function () { self.checkInduction(); });
            this.bindMenuAction('arg-bounded-check', function () { self.boundedCheck(); });
            this.bindMenuAction('arg-diagram', function () { self.diagramDomain(); });
            this.bindMenuAction('arg-weaken', function () { self.weakenInvariant(); });
            this.bindMenuAction('arg-save-invariant', function () { self.saveInvariant(); });

            // --- Concept Panel Menu Items (Conjecture, View) ---
            this.bindMenuAction('conj-undo', function () { self.doUndo(); });
            this.bindMenuAction('conj-redo', function () { self.doRedo(); });
            this.bindMenuAction('conj-pdr-step', function () { self.pdrStep(); });
            this.bindMenuAction('conj-concrete', function () { self.concreteStep(); });
            this.bindMenuAction('conj-gather', function () { self.gatherFacts(); });
            this.bindMenuAction('conj-cti-gather', function () { self.ctiConceptAction('cti_gather'); });
            this.bindMenuAction('conj-cti-minimize', function () { self.ctiConceptAction('cti_minimize'); });
            this.bindMenuAction('conj-cti-check-sufficient', function () { self.ctiConceptAction('cti_check_sufficient'); });
            this.bindMenuAction('conj-cti-check-inductive', function () { self.ctiConceptAction('cti_check_inductive'); });
            this.bindMenuAction('conj-cti-strengthen', function () { self.ctiConceptAction('cti_strengthen'); });
            this.bindMenuAction('conj-reverse', function () { self.reverseStep(); });
            this.bindMenuAction('conj-path-reach', function () { self.pathReach(); });
            this.bindMenuAction('conj-reach', function () { self.reachStep(); });
            this.bindMenuAction('conj-conjecture', function () { self.makeConjecture(); });
            this.bindMenuAction('conj-backtrack', function () { self.backtrack(); });
            this.bindMenuAction('conj-recalculate', function () { self.recalculateGraph(); });
            this.bindMenuAction('conj-diagram', function () { self.diagramDomain(); });
            this.bindMenuAction('conj-remember', function () { self.rememberGraph(); });
            this.bindMenuAction('conj-export', function () { self.exportConjecture(); });
            this.bindMenuAction('view-add-relation', function () { self.addRelationFromString(); });
        }

        if (!(window.__ivyVueBridge &&
            typeof window.__ivyVueBridge.globalInteractionsHandled === 'function' &&
            window.__ivyVueBridge.globalInteractionsHandled())) {
            // --- Click anywhere to dismiss context menu and dropdowns ---
            document.addEventListener('click', function (e) {
                self.controls.hideContextMenu();
                if (!e.target.closest('.dropdown')) {
                    self.closeAllDropdowns();
                }
            });

            // --- Escape key closes open dropdowns; Ctrl+S saves ---
            document.addEventListener('keydown', function (e) {
                if (e.key === 'Escape') {
                    self.closeAllDropdowns();
                }
                if ((e.ctrlKey || e.metaKey) && e.key === 's') {
                    e.preventDefault();
                    self.save();
                }
            });
        }

        this.attachGraphEventHandlers(this.argGraph, this.conceptGraph, 'sheet-1');
    }

    /**
     * Set up resizable dividers between ARG and concept panels.
     * Uses event delegation on the sheet area so it works for ALL tabs,
     * including dynamically created ones.
     */
    setupResizer() {
        var sheetArea = document.getElementById('sheet-area');
        if (!sheetArea) return;
        var self = this;
        var isDragging = false;
        var startX = 0;
        var startWidth = 0;
        var activeDivider = null;
        var activePanel = null;
        var activeContainer = null;

        // Delegation: any .divider inside a .sheet-main starts a drag
        sheetArea.addEventListener('mousedown', function (e) {
            var div = e.target;
            if (!div.classList.contains('divider')) return;
            var container = div.parentElement; // .sheet-main
            var panel = div.previousElementSibling; // ARG panel (left of divider)
            if (!container || !panel) return;

            isDragging = true;
            activeDivider = div;
            activePanel = panel;
            activeContainer = container;
            startX = e.clientX;
            startWidth = panel.offsetWidth;
            div.classList.add('active');
            document.body.style.cursor = 'col-resize';
            document.body.style.userSelect = 'none';
            var canvases = document.querySelectorAll('.graph-container');
            for (var i = 0; i < canvases.length; i++) canvases[i].style.pointerEvents = 'none';
            e.preventDefault();
        });

        document.addEventListener('mousemove', function (e) {
            if (!isDragging) return;
            var dx = e.clientX - startX;
            var newWidth = startWidth + dx;
            var containerWidth = activeContainer ? activeContainer.offsetWidth : 800;
            newWidth = Math.max(150, Math.min(newWidth, containerWidth - 200));
            if (!(activePanel.id === 'arg-panel' && self._setVueLayoutSize('setArgPanelWidth', newWidth))) {
                activePanel.style.flex = '0 0 ' + newWidth + 'px';
            }
            if (self.argGraph) self.argGraph.resize();
            if (self.conceptGraph) self.conceptGraph.resize();
        });

        document.addEventListener('mouseup', function () {
            if (isDragging) {
                isDragging = false;
                if (activeDivider) activeDivider.classList.remove('active');
                document.body.style.cursor = '';
                document.body.style.userSelect = '';
                var canvases = document.querySelectorAll('.graph-container');
                for (var i = 0; i < canvases.length; i++) canvases[i].style.pointerEvents = '';
                if (self.argGraph) self.argGraph.resize();
                if (self.conceptGraph) self.conceptGraph.resize();
                activeDivider = null;
                activePanel = null;
                activeContainer = null;
            }
        });
    }

    /**
     * Set up the resizable second divider between concept and state panels.
     */
    /**
     * Resizer for divider2: between sheet-left (ARG+Concept+Details) and state-panel.
     * Dragging left makes state panel wider; dragging right makes left wider.
     */
    setupResizer2() {
        var divider2 = document.getElementById('divider2');
        if (!divider2) return;
        var rightSection = document.getElementById('state-panel');
        var topRow = divider2.parentElement;
        var self = this;
        var isDragging = false;
        var startX = 0;
        var startWidth = 0;

        divider2.addEventListener('mousedown', function (e) {
            isDragging = true;
            startX = e.clientX;
            startWidth = rightSection.offsetWidth;
            divider2.classList.add('active');
            document.body.style.cursor = 'col-resize';
            document.body.style.userSelect = 'none';
            e.preventDefault();
        });

        document.addEventListener('mousemove', function (e) {
            if (!isDragging) return;
            var dx = startX - e.clientX; // drag left = right-section wider
            var newWidth = startWidth + dx;
            var maxW = topRow ? topRow.offsetWidth - 300 : 800;
            newWidth = Math.max(200, Math.min(newWidth, maxW));
            if (!self._setVueLayoutSize('setStatePanelWidth', newWidth)) {
                rightSection.style.flex = '0 0 ' + newWidth + 'px';
            }
            self.argGraph.resize();
            self.conceptGraph.resize();
        });

        document.addEventListener('mouseup', function () {
            if (isDragging) {
                isDragging = false;
                divider2.classList.remove('active');
                document.body.style.cursor = '';
                document.body.style.userSelect = '';
                self.argGraph.resize();
                self.conceptGraph.resize();
            }
        });
    }

    /**
     * Set up the resizable divider between State pane and Editor/Tutorial pane.
     */
    /**
     * Resizer for divider3: between sheet-area and editor-panel.
     * Dragging left makes editor wider; dragging right makes sheet area wider.
     */
    setupResizer3() {
        var divider3 = document.getElementById('divider3');
        if (!divider3) return;
        var sheetArea = document.getElementById('sheet-area');
        var editorPanel = document.getElementById('editor-panel');
        var topRow = document.getElementById('top-row');
        if (!sheetArea || !editorPanel || !topRow) return;
        var self = this;
        var isDragging = false;
        var startX = 0;
        var startWidth = 0;
        var minEditorWidth = 200;
        var minSheetAreaWidth = 400;

        divider3.addEventListener('mousedown', function (e) {
            isDragging = true;
            startX = e.clientX;
            startWidth = editorPanel.offsetWidth;
            sheetArea.style.flex = '1 1 auto';
            divider3.classList.add('active');
            document.body.style.cursor = 'col-resize';
            document.body.style.userSelect = 'none';
            e.preventDefault();
        });

        document.addEventListener('mousemove', function (e) {
            if (!isDragging) return;
            var dx = startX - e.clientX; // drag left = editor wider
            var newWidth = startWidth + dx;
            var dividerWidth = divider3.offsetWidth || 4;
            var maxW = Math.max(minEditorWidth, topRow.offsetWidth - dividerWidth - minSheetAreaWidth);
            newWidth = Math.max(minEditorWidth, Math.min(newWidth, maxW));
            if (!self._setVueLayoutSize('setEditorWidth', newWidth)) {
                editorPanel.style.flex = '0 0 ' + newWidth + 'px';
            }
            if (self.argGraph) self.argGraph.resize();
            if (self.conceptGraph) self.conceptGraph.resize();
            self._refreshEditorLayout();
        });

        document.addEventListener('mouseup', function () {
            if (isDragging) {
                isDragging = false;
                divider3.classList.remove('active');
                document.body.style.cursor = '';
                document.body.style.userSelect = '';
                if (self.argGraph) self.argGraph.resize();
                if (self.conceptGraph) self.conceptGraph.resize();
                self._refreshEditorLayout();
            }
        });
    }

    /**
     * Populate the state checkbox table (right pane) with edge/relation names.
     * Each row has checkboxes for: + (all_to_all), ? (unknown), - (none_to_none), T (transitive)
     * and the relation name.
     */
    /**
     * Set up the tutorial URL bar: Go button and Enter key navigate the iframe.
     */
    /**
     * Resizer for divider-h: horizontal divider between top row and tutorial BiB.
     * Dragging up makes tutorial taller; dragging down makes top row taller.
     */
    /**
     * Toggle the tutorial BiB panel visibility.
     */
    /**
     * Set up tab bar click handlers.
     * Matches Python tix.NoteBook: clicking a tab shows that sheet.
     */
    setupTabs() {
        var self = this;
        this._sheetCounter = 1;
        var tabBar = document.getElementById('tab-bar');
        if (!tabBar) return;
        if (window.__ivyVueBridge &&
            typeof window.__ivyVueBridge.tabClicksHandled === 'function' &&
            window.__ivyVueBridge.tabClicksHandled()) {
            return;
        }
        tabBar.addEventListener('click', function (e) {
            // Close button clicked?
            if (e.target.classList.contains('tab-close')) {
                var tab = e.target.parentElement;
                var sheetId = tab.getAttribute('data-sheet');
                self.removeSheet(sheetId);
                return;
            }
            var tab = e.target.closest('.sheet-tab');
            if (!tab) return;
            self.switchSheet(tab.getAttribute('data-sheet'));
        });
    }

    isValidSheetId(sheetId) {
        return /^[A-Za-z][A-Za-z0-9_-]*$/.test(String(sheetId || ''));
    }

    assertValidSheetId(sheetId) {
        if (!this.isValidSheetId(sheetId)) {
            throw new Error('invalid sheet id: ' + sheetId);
        }
        return sheetId;
    }

    sheetTab(sheetId) {
        var tabs = document.querySelectorAll('.sheet-tab');
        for (var i = 0; i < tabs.length; i++) {
            if (tabs[i].getAttribute('data-sheet') === sheetId) {
                return tabs[i];
            }
        }
        return null;
    }

    sheetExists(sheetId) {
        return !!((this.sheets && this.sheets[sheetId]) || document.getElementById(sheetId) || this.sheetTab(sheetId));
    }

    eventTraceRow(sheetId, address) {
        var sheet = document.getElementById(sheetId);
        if (!sheet) return null;
        var rows = sheet.querySelectorAll('.event-row');
        for (var i = 0; i < rows.length; i++) {
            if (rows[i].getAttribute('data-event-address') === address) {
                return rows[i];
            }
        }
        return null;
    }

    /**
     * Switch to a sheet by ID.
     */
    switchSheet(sheetId) {
        var bridge = window.__ivyVueBridge;
        var vueTabs = bridge && typeof bridge.activateSheetTab === 'function';
        if (vueTabs) {
            bridge.activateSheetTab(sheetId);
        }

        // Deactivate all tabs and sheets
        var tabs = vueTabs ? [] : document.querySelectorAll('.sheet-tab');
        var sheets = document.querySelectorAll('.sheet-content');
        for (var i = 0; i < tabs.length; i++) tabs[i].classList.remove('active');
        for (var i = 0; i < sheets.length; i++) {
            if (!vueTabs || sheets[i].__ivyVueRenderedSheet) {
                sheets[i].classList.remove('active');
            }
        }
        // Activate the target
        var tab = this.sheetTab(sheetId);
        var sheet = document.getElementById(sheetId);
        if (tab && !vueTabs) tab.classList.add('active');
        if (sheet && (!vueTabs || sheet.__ivyVueRenderedSheet)) sheet.classList.add('active');
        if (this.sheets && this.sheets[sheetId]) {
            this.activeSheetId = sheetId;
            if (this.sheets[sheetId].type !== 'events') {
                this.argGraph = this.sheets[sheetId].argGraph;
                this.conceptGraph = this.sheets[sheetId].conceptGraph;
                this.selectedArgNode = this.sheets[sheetId].selectedArgNode;
            }
        }
        // Resize graphs in the newly visible sheet
        if (this.argGraph) this.argGraph.resize();
        if (this.conceptGraph) this.conceptGraph.resize();
        if (bridge && typeof bridge.setActiveGraphSheet === 'function') {
            bridge.setActiveGraphSheet(sheetId);
        }
    }

    /**
     * Add a new sheet tab (matches Python ui_parent.add(art)).
     * Called on "Step into" / "Decompose" actions.
     * @param {string} [label] - Tab label (default: "Sheet N")
     * @returns {string} The new sheet ID
     */
    addSheet(label, preferredSheetId) {
        this._sheetCounter++;
        var sheetId = preferredSheetId || ('sheet-' + this._sheetCounter);
        while (!preferredSheetId && this.sheetExists(sheetId)) {
            this._sheetCounter++;
            sheetId = 'sheet-' + this._sheetCounter;
        }
        this.assertValidSheetId(sheetId);
        if (preferredSheetId && this.sheetExists(sheetId)) {
            throw new Error('duplicate sheet id: ' + sheetId);
        }
        var match = /^sheet-(\\d+)$/.exec(sheetId);
        if (match) {
            this._sheetCounter = Math.max(this._sheetCounter, parseInt(match[1], 10));
        }
        label = label || ('Sheet ' + this._sheetCounter);

        // Create tab button with close X (Sheet 1 never has X)
        if (window.__ivyVueBridge && typeof window.__ivyVueBridge.upsertSheetTab === 'function') {
            window.__ivyVueBridge.upsertSheetTab({ id: sheetId, label: label, closable: true, type: 'analysis' });
        } else {
            var tabBar = document.getElementById('tab-bar');
            var tabBtn = document.createElement('button');
            tabBtn.className = 'sheet-tab';
            tabBtn.setAttribute('data-sheet', sheetId);
            var labelSpan = document.createElement('span');
            labelSpan.textContent = label;
            tabBtn.appendChild(labelSpan);
            var closeBtn = document.createElement('span');
            closeBtn.className = 'tab-close';
            closeBtn.textContent = '\\u00D7'; // ×
            closeBtn.title = 'Close tab';
            tabBtn.appendChild(closeBtn);
            tabBar.appendChild(tabBtn);
        }

        // Create sheet content (clone structure from sheet-1)
        var newSheet = null;
        if (window.__ivyVueBridge && typeof window.__ivyVueBridge.createAnalysisSheetShell === 'function') {
            newSheet = window.__ivyVueBridge.createAnalysisSheetShell({ id: sheetId, counter: this._sheetCounter });
        }
        if (!newSheet) {
            var template = document.getElementById('sheet-1');
            newSheet = template.cloneNode(true);
            newSheet.id = sheetId;
            newSheet.classList.remove('active');
            // Clear graph containers (they'll be initialized fresh)
            var fallbackGraphs = newSheet.querySelectorAll('.graph-container');
            for (var g = 0; g < fallbackGraphs.length; g++) {
                fallbackGraphs[g].innerHTML = '';
                fallbackGraphs[g].id = fallbackGraphs[g].id + '-' + this._sheetCounter;
            }
            // Clear info panel
            var info = newSheet.querySelector('#info-content');
            if (info) {
                info.id = 'info-content-' + this._sheetCounter;
                info.textContent = 'Select a node or edge to see details';
            }
            var infoHeader = newSheet.querySelector('#info-header');
            if (infoHeader) infoHeader.id = 'info-header-' + this._sheetCounter;
            // Insert before the tutorial container
            var sheetArea = document.getElementById('sheet-area');
            sheetArea.appendChild(newSheet);
        }

        var graphs = newSheet.querySelectorAll('.graph-container');
        var graphIds = [];
        for (var i = 0; i < graphs.length; i++) {
            graphs[i].innerHTML = '';
            graphIds.push(graphs[i].id);
        }

        var argGraph = new IvyGraph(graphIds[0], ARG_STYLE);
        var conceptGraph = new IvyGraph(graphIds[1], CONCEPT_STYLE);
        argGraph.healthCheck();
        conceptGraph.healthCheck();
        this.registerSheet(sheetId, argGraph, conceptGraph);
        this.attachGraphEventHandlers(argGraph, conceptGraph, sheetId);

        // Switch to the new sheet
        this.switchSheet(sheetId);
        this.controls.setStatus('Opened: ' + label);
        return sheetId;
    }

    openARGSheet(label, argData, preferredSheetId) {
        var sheetId = this.addSheet(label, preferredSheetId);
        var sheetState = this.sheets && this.sheets[sheetId];
        if (sheetState && sheetState.argGraph && argData && argData.elements) {
            sheetState.argGraph.update(argData.elements, argData.positions);
        }
        return sheetId;
    }

    openEventTraceSheet(label, data, preferredSheetId) {
        data = data || {};
        var events = data.events || [];
        var patterns = data.patterns || [];
        this._sheetCounter = this._sheetCounter || 1;
        this._sheetCounter++;
        var sheetId = preferredSheetId || data.sheet_id || ('events-' + this._sheetCounter);
        this.assertValidSheetId(sheetId);
        var sheetArea = document.getElementById('sheet-area');
        if (!sheetArea) return '';

        var existingTab = this.sheetTab(sheetId);
        var existingSheet = document.getElementById(sheetId);
        var tabLabel = label || data.label || 'Events';
        var vueEventSheets = window.__ivyVueBridge && typeof window.__ivyVueBridge.upsertEventTraceSheet === 'function';
        if (window.__ivyVueBridge && typeof window.__ivyVueBridge.upsertSheetTab === 'function') {
            window.__ivyVueBridge.upsertSheetTab({ id: sheetId, label: tabLabel, closable: true, type: 'events' });
        } else {
            var tabBar = document.getElementById('tab-bar');
            if (!tabBar) return '';
            if (!existingTab) {
                var tabBtn = document.createElement('button');
                tabBtn.className = 'sheet-tab';
                tabBtn.setAttribute('data-sheet', sheetId);
                var labelSpan = document.createElement('span');
                labelSpan.textContent = tabLabel;
                tabBtn.appendChild(labelSpan);
                var closeBtn = document.createElement('span');
                closeBtn.className = 'tab-close';
                closeBtn.textContent = '\\u00D7';
                closeBtn.title = 'Close tab';
                tabBtn.appendChild(closeBtn);
                tabBar.appendChild(tabBtn);
            } else {
                var existingLabel = existingTab.querySelector('span');
                if (existingLabel) {
                    existingLabel.textContent = tabLabel;
                }
            }
        }

        var sheet = existingSheet;
        if (!sheet && !vueEventSheets) {
            sheet = document.createElement('div');
            sheet.id = sheetId;
            sheet.className = 'sheet-content event-sheet';
            sheetArea.appendChild(sheet);
        }

        this.sheets[sheetId] = {
            id: sheetId,
            type: 'events',
            events: events,
            patterns: patterns.slice(),
            selectedEventAddress: data.selected_address || null,
        };
        if (vueEventSheets) {
            window.__ivyVueBridge.upsertEventTraceSheet({
                id: sheetId,
                label: tabLabel,
                events: events,
                patterns: patterns,
                selectedEventAddress: data.selected_address || null,
            });
        } else {
            this.renderEventTraceSheet(sheetId);
        }
        this.switchSheet(sheetId);
        this.controls.setStatus('Opened: ' + (label || data.label || 'Events'));
        return sheetId;
    }

    async loadEventTraceFile(file) {
        if (!file) return null;
        this.controls.setStatus('Loading event trace...');
        try {
            var text = await this.readFileText(file);
            var result = await this.api.executeAction('events_parse', { content: text, filename: file.name || '' });
            this.openEventTraceSheet(file.name || result.label || 'Event trace', result || {}, result && result.sheet_id);
            this.controls.setStatus('Loaded event trace: ' + (file.name || 'trace'), 'success');
            return result;
        } catch (e) {
            this.controls.setStatus('Event trace load failed: ' + e.message, 'error');
            throw e;
        }
    }

    readFileText(file) {
        if (file && typeof file.text === 'function') {
            return file.text();
        }
        return new Promise(function (resolve, reject) {
            var reader = new FileReader();
            reader.onload = function () { resolve(String(reader.result || '')); };
            reader.onerror = function () { reject(reader.error || new Error('failed to read file')); };
            reader.readAsText(file);
        });
    }

    renderEventTraceSheet(sheetId) {
        var sheetState = this.sheets && this.sheets[sheetId];
        var sheet = document.getElementById(sheetId);
        if (!sheetState || !sheet) return;
        sheet.classList.add('event-sheet');
        sheet.innerHTML = [
            '<div class="event-viewer">',
            '  <div class="event-tree-panel panel">',
            '    <div class="panel-header">',
            '      <span class="panel-title">Events</span>',
            '      <button class="menu-btn event-filter-btn" type="button">Filter...</button>',
            '      <button class="menu-btn event-find-fwd-btn" type="button">&gt;&gt;</button>',
            '      <button class="menu-btn event-find-rev-btn" type="button">&lt;&lt;</button>',
            '    </div>',
            '    <div class="event-tree" data-event-tree="' + sheetId + '"></div>',
            '  </div>',
            '  <div class="event-pattern-panel panel">',
            '    <div class="panel-header"><span class="panel-title">Patterns</span></div>',
            '    <select class="event-pattern-list" size="8"></select>',
            '    <div class="event-pattern-buttons">',
            '      <button class="menu-btn event-pattern-rev" type="button">&lt;&lt;</button>',
            '      <button class="menu-btn event-pattern-fwd" type="button">&gt;&gt;</button>',
            '      <button class="menu-btn event-pattern-add" type="button">+</button>',
            '      <button class="menu-btn event-pattern-remove" type="button">-</button>',
            '    </div>',
            '    <div class="event-pattern-buttons">',
            '      <button class="menu-btn event-pattern-save" type="button">Save</button>',
            '      <button class="menu-btn event-pattern-load" type="button">Load</button>',
            '      <button class="menu-btn event-pattern-clear" type="button">Clear</button>',
            '    </div>',
            '  </div>',
            '</div>',
        ].join('');

        var tree = sheet.querySelector('.event-tree');
        this.renderEventTree(tree, sheetState.events || [], sheetId, '');
        this.renderEventPatternList(sheetId);
        this.attachEventTraceHandlers(sheetId);
        if (sheetState.selectedEventAddress) {
            this.selectEventTraceRow(sheetId, sheetState.selectedEventAddress);
        }
    }

    renderEventTree(container, events, sheetId, prefix) {
        if (!container) return;
        container.innerHTML = '';
        var list = document.createElement('ul');
        list.className = 'event-tree-list';
        for (var i = 0; i < (events || []).length; i++) {
            list.appendChild(this.renderEventTreeNode(events[i], sheetId, prefix === '' ? String(i) : prefix + '/' + i));
        }
        container.appendChild(list);
    }

    renderEventTreeNode(ev, sheetId, fallbackAddress) {
        ev = ev || {};
        var address = ev.address || fallbackAddress;
        ev.address = address;
        var li = document.createElement('li');
        li.className = 'event-tree-node';
        li.setAttribute('data-event-node', address);

        var row = document.createElement('div');
        row.className = 'event-row';
        row.setAttribute('data-event-address', address);

        var hasSubs = ev.subs && ev.subs.length > 0;
        var toggle = document.createElement('button');
        toggle.className = 'event-toggle';
        toggle.type = 'button';
        toggle.textContent = hasSubs ? '+' : '';
        toggle.disabled = !hasSubs;
        if (hasSubs) toggle.setAttribute('data-event-toggle', address);
        row.appendChild(toggle);

        var text = document.createElement('span');
        text.className = 'event-text';
        text.textContent = ev.text || '';
        row.appendChild(text);
        li.appendChild(row);

        row.addEventListener('click', this.selectEventTraceRow.bind(this, sheetId, address));
        if (hasSubs) {
            toggle.addEventListener('click', function (e) {
                e.stopPropagation();
                this.toggleEventTraceNode(sheetId, address);
            }.bind(this));
        }
        return li;
    }

    attachEventTraceHandlers(sheetId) {
        var sheet = document.getElementById(sheetId);
        if (!sheet) return;
        var self = this;
        var filter = sheet.querySelector('.event-filter-btn');
        if (filter) filter.addEventListener('click', async function () {
            var pat = await self.entryDialog('Filter events', 'Pattern:', '', { okLabel: 'Filter' });
            if (pat !== null) await self.filterEventTrace(pat);
        });
        var findFwd = sheet.querySelector('.event-find-fwd-btn');
        if (findFwd) findFwd.addEventListener('click', async function () {
            var pat = await self.entryDialog('Find event', 'Pattern:', '', { okLabel: 'Find' });
            if (pat !== null) await self.findEventTrace(pat, false);
        });
        var findRev = sheet.querySelector('.event-find-rev-btn');
        if (findRev) findRev.addEventListener('click', async function () {
            var pat = await self.entryDialog('Find event', 'Pattern:', '', { okLabel: 'Find' });
            if (pat !== null) await self.findEventTrace(pat, true);
        });
        var patRev = sheet.querySelector('.event-pattern-rev');
        if (patRev) patRev.addEventListener('click', function () {
            var pat = self.selectedEventPattern(sheetId);
            if (pat) self.findEventTrace(pat, true);
        });
        var patFwd = sheet.querySelector('.event-pattern-fwd');
        if (patFwd) patFwd.addEventListener('click', function () {
            var pat = self.selectedEventPattern(sheetId);
            if (pat) self.findEventTrace(pat, false);
        });
        var patAdd = sheet.querySelector('.event-pattern-add');
        if (patAdd) patAdd.addEventListener('click', async function () {
            var pat = await self.entryDialog('Add pattern', 'Pattern:', '', { okLabel: 'Add' });
            if (pat !== null && pat !== '') await self.addEventPattern(sheetId, pat);
        });
        var patRemove = sheet.querySelector('.event-pattern-remove');
        if (patRemove) patRemove.addEventListener('click', function () {
            self.removeSelectedEventPattern(sheetId);
        });
        var patSave = sheet.querySelector('.event-pattern-save');
        if (patSave) patSave.addEventListener('click', function () {
            self.saveEventPatterns(sheetId);
        });
        var patLoad = sheet.querySelector('.event-pattern-load');
        if (patLoad) patLoad.addEventListener('click', async function () {
            var text = await self.textDialog('Load patterns', 'Paste patterns:', '', { okLabel: 'Load' });
            if (text !== null) await self.loadEventPatterns(sheetId, text);
        });
        var patClear = sheet.querySelector('.event-pattern-clear');
        if (patClear) patClear.addEventListener('click', function () {
            self.clearEventPatterns(sheetId);
        });
    }

    toggleEventTraceNode(sheetId, address) {
        if (window.__ivyVueBridge && typeof window.__ivyVueBridge.setEventTraceExpanded === 'function') {
            var expanded = false;
            if (typeof window.__ivyVueBridge.isEventTraceExpanded === 'function') {
                expanded = !!window.__ivyVueBridge.isEventTraceExpanded(sheetId, address);
            }
            window.__ivyVueBridge.setEventTraceExpanded(sheetId, address, !expanded);
            return;
        }
        var row = this.eventTraceRow(sheetId, address);
        var li = row ? row.closest('.event-tree-node') : null;
        var sheetState = this.sheets && this.sheets[sheetId];
        if (!li || !sheetState) return;
        var existing = li.querySelector(':scope > ul.event-tree-list');
        var toggle = row.querySelector('.event-toggle');
        if (existing) {
            existing.remove();
            if (toggle) toggle.textContent = '+';
            return;
        }
        var ev = this.lookupEventTrace(sheetState.events, address);
        if (!ev || !ev.subs || ev.subs.length === 0) return;
        var list = document.createElement('ul');
        list.className = 'event-tree-list';
        for (var i = 0; i < ev.subs.length; i++) {
            list.appendChild(this.renderEventTreeNode(ev.subs[i], sheetId, address + '/' + i));
        }
        li.appendChild(list);
        if (toggle) toggle.textContent = '-';
    }

    lookupEventTrace(events, address) {
        if (!address && address !== '0') return null;
        var parts = String(address).split('/');
        var current = null;
        var list = events || [];
        for (var i = 0; i < parts.length; i++) {
            var idx = Number(parts[i]);
            if (!Number.isInteger(idx) || idx < 0 || idx >= list.length) return null;
            current = list[idx];
            list = current.subs || [];
        }
        return current;
    }

    uncoverEventTraceAddress(sheetId, address) {
        var parts = String(address || '').split('/');
        var prefix = '';
        for (var i = 0; i < parts.length - 1; i++) {
            prefix = prefix === '' ? parts[i] : prefix + '/' + parts[i];
            if (!this.eventTraceRow(sheetId, prefix)) break;
            if (!this.eventTraceRow(sheetId, prefix + '/' + parts[i + 1])) {
                this.toggleEventTraceNode(sheetId, prefix);
            }
        }
    }

    selectEventTraceRow(sheetId, address) {
        var sheetState = this.sheets && this.sheets[sheetId];
        if (window.__ivyVueBridge && typeof window.__ivyVueBridge.selectEventTraceRow === 'function') {
            if (sheetState) sheetState.selectedEventAddress = address;
            window.__ivyVueBridge.selectEventTraceRow(sheetId, address);
            return;
        }
        this.uncoverEventTraceAddress(sheetId, address);
        var sheet = document.getElementById(sheetId);
        if (!sheetState || !sheet) return;
        var rows = sheet.querySelectorAll('.event-row.selected');
        for (var i = 0; i < rows.length; i++) rows[i].classList.remove('selected');
        var row = this.eventTraceRow(sheetId, address);
        if (row) {
            row.classList.add('selected');
            if (typeof row.scrollIntoView === 'function') {
                row.scrollIntoView({ block: 'nearest' });
            }
        }
        sheetState.selectedEventAddress = address;
    }

    activeEventSheet() {
        var sheet = this.sheets && this.sheets[this.activeSheetId];
        return sheet && sheet.type === 'events' ? sheet : null;
    }

    async filterEventTrace(pattern) {
        var sheet = this.activeEventSheet();
        if (!sheet) {
            this.controls.setStatus('No event sheet selected', 'error');
            return null;
        }
        if (sheet.visualOnly) {
            this.controls.setStatus(this.visualOnlyMessage('events'), 'warning');
            return null;
        }
        try {
            var result = await this.api.executeAction('events_filter', {
                sheet_id: sheet.id,
                pattern: pattern,
            });
            var label = (result && result.label) || 'Filtered events';
            this.openEventTraceSheet(label, result || {}, result && result.sheet_id);
            return result;
        } catch (e) {
            this.controls.setStatus('Filter failed: ' + e.message, 'error');
            return null;
        }
    }

    async findEventTrace(pattern, reverse) {
        var sheet = this.activeEventSheet();
        if (!sheet) {
            this.controls.setStatus('No event sheet selected', 'error');
            return null;
        }
        if (sheet.visualOnly) {
            this.controls.setStatus(this.visualOnlyMessage('events'), 'warning');
            return null;
        }
        var args = {
            sheet_id: sheet.id,
            pattern: pattern,
            reverse: !!reverse,
            anchor: sheet.selectedEventAddress || '',
        };
        var result;
        try {
            result = await this.api.executeAction('events_find', args);
        } catch (e) {
            this.controls.setStatus('Find failed: ' + e.message, 'error');
            return null;
        }
        if (!result || !result.address) {
            this.controls.setStatus('Pattern not found', 'error');
            return result;
        }
        this.selectEventTraceRow(sheet.id, result.address);
        return result;
    }

    applyEventPatternResult(sheetId, result, fallbackPatterns) {
        var sheet = this.sheets && this.sheets[sheetId];
        if (!sheet) return;
        if (result && Array.isArray(result.patterns)) {
            sheet.patterns = result.patterns.slice();
        } else if (fallbackPatterns) {
            sheet.patterns = fallbackPatterns.slice();
        }
        if (window.__ivyVueBridge && typeof window.__ivyVueBridge.updateEventPatterns === 'function') {
            window.__ivyVueBridge.updateEventPatterns(sheetId, sheet.patterns || []);
            return;
        }
        this.renderEventPatternList(sheetId);
    }

    renderEventPatternList(sheetId) {
        var sheet = document.getElementById(sheetId);
        var sheetState = this.sheets && this.sheets[sheetId];
        if (window.__ivyVueBridge && typeof window.__ivyVueBridge.updateEventPatterns === 'function') {
            if (sheetState) {
                window.__ivyVueBridge.updateEventPatterns(sheetId, sheetState.patterns || []);
            }
            return;
        }
        var select = sheet ? sheet.querySelector('.event-pattern-list') : null;
        if (!select || !sheetState) return;
        select.innerHTML = '';
        for (var i = 0; i < (sheetState.patterns || []).length; i++) {
            var option = document.createElement('option');
            option.value = sheetState.patterns[i];
            option.textContent = sheetState.patterns[i];
            select.appendChild(option);
        }
    }

    selectedEventPattern(sheetId) {
        if (window.__ivyVueBridge && typeof window.__ivyVueBridge.getSelectedEventPattern === 'function') {
            return window.__ivyVueBridge.getSelectedEventPattern(sheetId) || '';
        }
        var sheet = document.getElementById(sheetId);
        var select = sheet ? sheet.querySelector('.event-pattern-list') : null;
        return select && select.value ? select.value : '';
    }

    async addEventPattern(sheetId, pattern) {
        var sheet = this.sheets && this.sheets[sheetId];
        if (!sheet) return;
        if (this.api && this.api.executeAction && !sheet.visualOnly) {
            var result = await this.api.executeAction('events_add_pattern', { sheet_id: sheetId, pattern: pattern });
            this.applyEventPatternResult(sheetId, result);
            return;
        }
        sheet.patterns = (sheet.patterns || []).concat([pattern]);
        this.renderEventPatternList(sheetId);
    }

    async removeSelectedEventPattern(sheetId) {
        var sheet = this.sheets && this.sheets[sheetId];
        var sheetEl = document.getElementById(sheetId);
        var select = sheetEl ? sheetEl.querySelector('.event-pattern-list') : null;
        var idx = -1;
        if (window.__ivyVueBridge && typeof window.__ivyVueBridge.getSelectedEventPatternIndex === 'function') {
            idx = window.__ivyVueBridge.getSelectedEventPatternIndex(sheetId);
        } else if (select) {
            idx = select.selectedIndex;
        }
        if (!sheet || idx < 0) return;
        if (this.api && this.api.executeAction && !sheet.visualOnly) {
            var result = await this.api.executeAction('events_remove_pattern', { sheet_id: sheetId, index: idx });
            this.applyEventPatternResult(sheetId, result);
            return;
        }
        sheet.patterns.splice(idx, 1);
        this.renderEventPatternList(sheetId);
    }

    async clearEventPatterns(sheetId) {
        var sheet = this.sheets && this.sheets[sheetId];
        if (!sheet) return;
        if (this.api && this.api.executeAction && !sheet.visualOnly) {
            var result = await this.api.executeAction('events_clear_patterns', { sheet_id: sheetId });
            this.applyEventPatternResult(sheetId, result, []);
            return;
        }
        sheet.patterns = [];
        this.renderEventPatternList(sheetId);
    }

    async loadEventPatterns(sheetId, text) {
        var sheet = this.sheets && this.sheets[sheetId];
        if (!sheet) return;
        if (this.api && this.api.executeAction && !sheet.visualOnly) {
            var result = await this.api.executeAction('events_load_patterns', { sheet_id: sheetId, patterns: text });
            this.applyEventPatternResult(sheetId, result);
            return;
        }
        var patterns = String(text || '').split(/\\r?\\n/).map(function (line) { return line.trim(); }).filter(Boolean);
        sheet.patterns = sheet.patterns.concat(patterns);
        this.renderEventPatternList(sheetId);
    }

    async saveEventPatterns(sheetId) {
        var sheet = this.sheets && this.sheets[sheetId];
        if (!sheet) return '';
        var content = (sheet.patterns || []).join('\\n');
        if (content !== '') content += '\\n';
        if (this.api && this.api.executeAction && !sheet.visualOnly) {
            var result = await this.api.executeAction('events_save_patterns', { sheet_id: sheetId });
            if (result && typeof result.content === 'string') {
                content = result.content;
            }
        }
        this.downloadTextFile('event_patterns.pats', content, 'text/plain');
        return content;
    }

    /**
     * Remove a sheet tab and its content.
     * If the removed sheet was active, switch to Sheet 1.
     * Sheet 1 cannot be removed.
     */
    removeSheet(sheetId) {
        if (sheetId === 'sheet-1') return; // never remove Sheet 1

        var tab = this.sheetTab(sheetId);
        var sheet = document.getElementById(sheetId);
        var wasActive = this.activeSheetId === sheetId || (tab && tab.classList.contains('active'));
        var sheetState = this.sheets && this.sheets[sheetId];
        var bridge = window.__ivyVueBridge;
        var vueTabs = bridge && typeof bridge.removeSheetTab === 'function';
        var vueOwnsSheetDom = vueTabs && sheetState && sheetState.type === 'events';

        if (tab && !vueTabs) tab.remove();
        if (bridge && typeof bridge.removeRenderedSheet === 'function') {
            bridge.removeRenderedSheet(sheetId);
        }
        if (sheet && !vueOwnsSheetDom) sheet.remove();
        if (this.sheets) {
            delete this.sheets[sheetId];
        }
        if (vueTabs) {
            bridge.removeSheetTab(sheetId);
        }

        // If the closed tab was active, switch to Sheet 1
        if (wasActive) {
            this.switchSheet('sheet-1');
        }
    }

    toggleTutorial(flash) {
        var bridge = window.__ivyVueBridge;
        if (bridge && typeof bridge.setTutorialVisible === 'function') {
            var tutorialForState = document.getElementById('tutorial-container');
            var visible = tutorialForState ? tutorialForState.style.display !== 'none' : true;
            if (typeof bridge.isTutorialVisible === 'function') {
                visible = !!bridge.isTutorialVisible();
            }
            var nextVisible = !visible;
            bridge.setTutorialVisible(nextVisible);
            if (!nextVisible && flash) {
                if (typeof bridge.flashTutorialButton === 'function') {
                    bridge.flashTutorialButton(1200);
                } else {
                    var vueFallbackButton = document.getElementById('btn-toggle-tutorial');
                    if (vueFallbackButton) {
                        vueFallbackButton.classList.add('btn-flash');
                        setTimeout(function () { vueFallbackButton.classList.remove('btn-flash'); }, 1200);
                    }
                }
            }
            this._refreshLayoutAfterVuePatch();
            return;
        }

        var tutorial = document.getElementById('tutorial-container');
        var dividerH = document.getElementById('divider-h');
        var btn = document.getElementById('btn-toggle-tutorial');
        if (!tutorial || !btn) return;

        if (tutorial.style.display === 'none') {
            // Show
            tutorial.style.display = '';
            if (dividerH) dividerH.style.display = '';
            btn.textContent = 'Hide Tutorial';
        } else {
            // Hide
            tutorial.style.display = 'none';
            if (dividerH) dividerH.style.display = 'none';
            btn.textContent = 'Show Tutorial';
            // Flash the button to alert the user where to find it again
            if (flash) {
                btn.classList.add('btn-flash');
                setTimeout(function () { btn.classList.remove('btn-flash'); }, 1200);
            }
        }
        // Resize graphs to fill the reclaimed/reduced space
        if (this.argGraph) this.argGraph.resize();
        if (this.conceptGraph) this.conceptGraph.resize();
        this._refreshEditorLayout();
    }

    _refreshGraphsAndEditorLayout() {
        if (this.argGraph) this.argGraph.resize();
        if (this.conceptGraph) this.conceptGraph.resize();
        this._refreshEditorLayout();
    }

    _refreshLayoutAfterVuePatch() {
        var self = this;
        var refresh = function () {
            self._refreshGraphsAndEditorLayout();
        };
        var bridge = window.__ivyVueBridge;
        if (bridge && typeof bridge.afterLayoutSettled === 'function') {
            bridge.afterLayoutSettled(function () {
                refresh();
                setTimeout(refresh, 60);
            });
            return;
        }
        if (typeof window.requestAnimationFrame === 'function') {
            window.requestAnimationFrame(function () {
                window.requestAnimationFrame(refresh);
            });
        } else {
            setTimeout(refresh, 0);
        }
    }

    setupResizerH() {
        var dividerH = document.getElementById('divider-h');
        if (!dividerH) return;
        var tutorial = document.getElementById('tutorial-container');
        var outerContainer = document.getElementById('outer-container');
        var iframe = document.getElementById('tutorial-iframe');
        var self = this;
        var isDragging = false;
        var startY = 0;
        var startHeight = 0;

        dividerH.addEventListener('mousedown', function (e) {
            isDragging = true;
            startY = e.clientY;
            startHeight = tutorial.offsetHeight;
            dividerH.classList.add('active');
            document.body.style.cursor = 'row-resize';
            document.body.style.userSelect = 'none';
            // Block iframe from stealing mouse events during drag
            if (iframe) iframe.style.pointerEvents = 'none';
            e.preventDefault();
        });

        document.addEventListener('mousemove', function (e) {
            if (!isDragging) return;
            var dy = startY - e.clientY;
            var newHeight = startHeight + dy;
            var maxH = outerContainer ? outerContainer.offsetHeight - 100 : 600;
            newHeight = Math.max(80, Math.min(newHeight, maxH));
            if (!self._setVueLayoutSize('setTutorialHeight', newHeight)) {
                tutorial.style.flex = '0 0 ' + newHeight + 'px';
            }
            self.argGraph.resize();
            self.conceptGraph.resize();
        });

        document.addEventListener('mouseup', function () {
            if (isDragging) {
                isDragging = false;
                dividerH.classList.remove('active');
                document.body.style.cursor = '';
                document.body.style.userSelect = '';
                if (iframe) iframe.style.pointerEvents = '';
                self.argGraph.resize();
                self.conceptGraph.resize();
            }
        });
    }

    setupDetailsResizer() {
        var sheetArea = document.getElementById('sheet-area');
        if (!sheetArea) return;
        var self = this;
        var isDragging = false;
        var startY = 0;
        var startHeight = 0;
        var activeHeader = null;
        var activePanel = null;
        var activeSheetLeft = null;
        var activeSheetMain = null;

        sheetArea.addEventListener('mousedown', function (e) {
            var header = e.target.closest('.info-header, #info-header');
            if (!header || !sheetArea.contains(header)) return;
            var panel = header.closest('.info-panel') || header.parentElement;
            var sheetLeft = panel ? panel.closest('.sheet-left') : null;
            if (!panel || !sheetLeft) return;

            isDragging = true;
            startY = e.clientY;
            startHeight = panel.offsetHeight;
            activeHeader = header;
            activePanel = panel;
            activeSheetLeft = sheetLeft;
            activeSheetMain = sheetLeft.querySelector('.sheet-main');
            header.classList.add('active');
            document.body.style.cursor = 'row-resize';
            document.body.style.userSelect = 'none';
            var graphs = sheetLeft.querySelectorAll('.graph-container');
            for (var i = 0; i < graphs.length; i++) graphs[i].style.pointerEvents = 'none';
            e.preventDefault();
        });

        document.addEventListener('mousemove', function (e) {
            if (!isDragging || !activePanel || !activeSheetLeft) return;
            var dy = startY - e.clientY;
            var minDetailsHeight = 72;
            var minMainHeight = self._detailsResizerMinimumMainHeight(activeSheetLeft);
            var maxHeight = Math.max(minDetailsHeight, activeSheetLeft.offsetHeight - minMainHeight);
            var newHeight = Math.max(minDetailsHeight, Math.min(startHeight + dy, maxHeight));
            if (activeSheetMain) {
                activeSheetMain.style.minHeight = minMainHeight + 'px';
            }
            if (!(activePanel.id === 'info-panel' && self._setVueLayoutSize('setDetailsHeight', newHeight))) {
                activePanel.style.flex = '0 0 ' + newHeight + 'px';
                activePanel.style.height = newHeight + 'px';
            }
            if (self.argGraph) self.argGraph.resize();
            if (self.conceptGraph) self.conceptGraph.resize();
        });

        document.addEventListener('mouseup', function () {
            if (!isDragging) return;
            isDragging = false;
            if (activeHeader) activeHeader.classList.remove('active');
            document.body.style.cursor = '';
            document.body.style.userSelect = '';
            if (activeSheetLeft) {
                var graphs = activeSheetLeft.querySelectorAll('.graph-container');
                for (var i = 0; i < graphs.length; i++) graphs[i].style.pointerEvents = '';
            }
            if (self.argGraph) self.argGraph.resize();
            if (self.conceptGraph) self.conceptGraph.resize();
            activeHeader = null;
            activePanel = null;
            activeSheetLeft = null;
            activeSheetMain = null;
        });
    }

    _detailsResizerMinimumMainHeight(sheetLeft) {
        var fallback = 44;
        if (!sheetLeft) return fallback;
        var sheetMain = sheetLeft.querySelector('.sheet-main');
        if (!sheetMain) return fallback;
        var headers = sheetMain.querySelectorAll('.panel-header');
        var height = fallback;
        for (var i = 0; i < headers.length; i++) {
            height = Math.max(height, headers[i].offsetHeight || 0);
        }
        return height;
    }

    setupTutorialUrlBar() {
        if (window.__ivyVueBridge &&
            typeof window.__ivyVueBridge.tutorialUrlBarHandled === 'function' &&
            window.__ivyVueBridge.tutorialUrlBarHandled()) {
            return;
        }
        var urlInput = document.getElementById('tutorial-url');
        var iframe = document.getElementById('tutorial-iframe');
        var backBtn = document.getElementById('tutorial-back');
        var fwdBtn = document.getElementById('tutorial-fwd');
        var reloadBtn = document.getElementById('tutorial-reload');
        var closeBtn = document.getElementById('tutorial-close');
        if (!urlInput || !iframe) return;

        // Close button: hide tutorial and flash the "Show Tutorial" button
        var self = this;
        if (closeBtn) {
            closeBtn.addEventListener('click', function () {
                self.toggleTutorial(true);
            });
        }

        // Track navigation history.
        var history = [urlInput.value.trim()];
        var historyIdx = 0;

        function navigateTo(url, forceReload) {
            if (!url) return;
            if (!url.match(/^https?:\\/\\//) && !url.startsWith('/')) {
                url = 'https://' + url;
            }
            if (forceReload) {
                // Force reload: about:blank trick bypasses browser cache
                iframe.src = 'about:blank';
                setTimeout(function () { iframe.src = url; }, 0);
            } else {
                // Normal navigation: browser cache can serve the page offline
                iframe.src = url;
            }
            if (historyIdx < history.length - 1) {
                history = history.slice(0, historyIdx + 1);
            }
            history.push(url);
            historyIdx = history.length - 1;
            urlInput.value = url;
            updateNavButtons();
        }

        function updateNavButtons() {
            if (backBtn) backBtn.disabled = (historyIdx <= 0);
            if (fwdBtn) fwdBtn.disabled = (historyIdx >= history.length - 1);
        }

        // Enter key navigates
        urlInput.addEventListener('keydown', function (e) {
            if (e.key === 'Enter') {
                e.preventDefault();
                navigateTo(urlInput.value.trim());
            }
        });

        // Back button
        if (backBtn) {
            backBtn.addEventListener('click', function () {
                if (historyIdx > 0) {
                    historyIdx--;
                    var url = history[historyIdx];
                    urlInput.value = url;
                    iframe.src = url;
                    updateNavButtons();
                }
            });
        }

        // Forward button
        if (fwdBtn) {
            fwdBtn.addEventListener('click', function () {
                if (historyIdx < history.length - 1) {
                    historyIdx++;
                    var url = history[historyIdx];
                    urlInput.value = url;
                    iframe.src = url;
                    updateNavButtons();
                }
            });
        }

        // Reload button — force reload (bypasses cache, for when server is back)
        if (reloadBtn) {
            reloadBtn.addEventListener('click', function () {
                var url = history[historyIdx];
                if (url) {
                    navigateTo(url, true);
                }
            });
        }

        // Update URL bar when iframe navigates, and cache successful pages.
        iframe.addEventListener('load', function () {
            try {
                var newUrl = iframe.contentWindow.location.href;
                // Convert full URL to path for cleaner display
                if (newUrl && newUrl !== 'about:blank') {
                    try {
                        var u = new URL(newUrl);
                        newUrl = u.pathname + u.search + u.hash;
                    } catch (e) {}
                    if (history[historyIdx] !== newUrl) {
                        if (historyIdx < history.length - 1) {
                            history = history.slice(0, historyIdx + 1);
                        }
                        history.push(newUrl);
                        historyIdx = history.length - 1;
                    }
                    urlInput.value = newUrl;
                }
            } catch (e) {
                // Cross-origin or error page — don't cache
            }
            updateNavButtons();
        });

        updateNavButtons();
    }

    populateStateCheckboxes(conceptData) {
        // Store concept data for node label rendering
        this._lastConceptData = conceptData;
        var tbody = document.getElementById('state-checkbox-body');
        this._hydrateBackendToggleState(conceptData);

        // Use the relations list from the server (edges + node_labels).
        // This matches Python's Graph.relation_ids.
        var names = [];
        if (conceptData && conceptData.relations) {
            names = conceptData.relations.slice().sort();
        }
        var self = this;

        var appendLegacyRows = function () {
            if (!tbody) return;
            tbody.innerHTML = '';
            for (var i = 0; i < names.length; i++) {
                (function(name) {
                    var tr = document.createElement('tr');

                    // + checkbox (all_to_all)
                    var td1 = document.createElement('td');
                    var cb1 = document.createElement('input');
                    cb1.type = 'checkbox';
                    cb1.name = name;
                    cb1.value = 'all_to_all';
                    cb1.title = 'Show definite edges (' + name + ')';
                    cb1.checked = self._toggleChecked(name, 'all_to_all');
                    cb1.addEventListener('change', function() {
                        self.onEdgeToggle(name, 'all_to_all', cb1.checked);
                    });
                    td1.appendChild(cb1);
                    tr.appendChild(td1);

                    // ? checkbox (unknown)
                    var td2 = document.createElement('td');
                    var cb2 = document.createElement('input');
                    cb2.type = 'checkbox';
                    cb2.name = name;
                    cb2.value = 'edge_unknown';
                    cb2.title = 'Show unknown edges (' + name + ')';
                    cb2.checked = self._toggleChecked(name, 'edge_unknown');
                    cb2.addEventListener('change', function() {
                        self.onEdgeToggle(name, 'edge_unknown', cb2.checked);
                    });
                    td2.appendChild(cb2);
                    tr.appendChild(td2);

                    // - checkbox (none_to_none)
                    var td3 = document.createElement('td');
                    var cb3 = document.createElement('input');
                    cb3.type = 'checkbox';
                    cb3.name = name;
                    cb3.value = 'none_to_none';
                    cb3.title = 'Show absent edges (' + name + ')';
                    cb3.checked = self._toggleChecked(name, 'none_to_none');
                    cb3.addEventListener('change', function() {
                        self.onEdgeToggle(name, 'none_to_none', cb3.checked);
                    });
                    td3.appendChild(cb3);
                    tr.appendChild(td3);

                    // T checkbox (transitive reduction)
                    var td4 = document.createElement('td');
                    var cb4 = document.createElement('input');
                    cb4.type = 'checkbox';
                    cb4.name = name;
                    cb4.value = 'transitive';
                    cb4.title = 'Transitive reduction (' + name + ')';
                    cb4.checked = self._toggleChecked(name, 'transitive');
                    cb4.addEventListener('change', function() {
                        self.onEdgeToggle(name, 'transitive', cb4.checked);
                    });
                    td4.appendChild(cb4);
                    tr.appendChild(td4);

                    // Name column
                    var td5 = document.createElement('td');
                    td5.className = 'name-col';
                    var a = document.createElement('a');
                    a.textContent = name;
                    a.href = '#';
                    a.addEventListener('click', function(e) {
                        e.preventDefault();
                        // Clicking the name could highlight related edges
                    });
                    td5.appendChild(a);
                    tr.appendChild(td5);

                    tbody.appendChild(tr);
                })(names[i]);
            }

            // If no edges found, show a placeholder
            if (names.length === 0 && conceptData) {
                var tr = document.createElement('tr');
                var td = document.createElement('td');
                td.colSpan = 5;
                td.style.color = '#666';
                td.style.fontStyle = 'italic';
                td.textContent = 'No relations loaded';
                tr.appendChild(td);
                tbody.appendChild(tr);
            }
        };

        var vueRows = names.map(function (name) {
            return {
                name: name,
                checked: {
                    all_to_all: self._toggleChecked(name, 'all_to_all'),
                    edge_unknown: self._toggleChecked(name, 'edge_unknown'),
                    none_to_none: self._toggleChecked(name, 'none_to_none'),
                    transitive: self._toggleChecked(name, 'transitive')
                }
            };
        });
        if (window.__ivyVueBridge && typeof window.__ivyVueBridge.updateStateRelations === 'function') {
            window.__ivyVueBridge.updateStateRelations(vueRows, function (name, displayClass, checked) {
                self.onEdgeToggle(name, displayClass, checked);
            });
            this._applyEdgeVisibility();
            this._applyNodeLabels();
            this.populateConstraintFacts(conceptData);
            return;
        }

        if (!tbody) return;
        appendLegacyRows();

        this._applyEdgeVisibility();
        this._applyNodeLabels();
        this.populateConstraintFacts(conceptData);
    }

    populateConstraintFacts(conceptData) {
        var facts = (conceptData && Array.isArray(conceptData.facts)) ? conceptData.facts : [];
        var self = this;
        if (window.__ivyVueBridge && typeof window.__ivyVueBridge.updateConstraintFacts === 'function') {
            window.__ivyVueBridge.updateConstraintFacts(facts, async function (index, selected) {
                await self.api.executeAction('set_fact_selection', {
                    index: index,
                    selected: selected,
                });
            });
            return;
        }
        var info = document.getElementById('info-content');
        if (!info) return;
        info.innerHTML = '';
        if (facts.length === 0) {
            info.textContent = 'Select a node or edge to see details';
            return;
        }

        var title = document.createElement('div');
        title.className = 'constraint-facts-title';
        title.textContent = 'Constraints:';
        info.appendChild(title);

        facts.forEach(function(fact, offset) {
            var index = typeof fact.index === 'number' ? fact.index : offset;
            var row = document.createElement('button');
            row.type = 'button';
            row.className = 'constraint-fact';
            row.setAttribute('data-constraint-fact', String(index));
            row.setAttribute('aria-pressed', fact.selected ? 'true' : 'false');
            row.textContent = fact.text || '';
            if (!fact.selected) {
                row.classList.add('inactive');
            }
            row.addEventListener('click', async function() {
                var selected = row.classList.contains('inactive');
                row.classList.toggle('inactive', !selected);
                row.setAttribute('aria-pressed', selected ? 'true' : 'false');
                try {
                    await self.api.executeAction('set_fact_selection', {
                        index: index,
                        selected: selected,
                    });
                } catch (e) {
                    row.classList.toggle('inactive', selected);
                    row.setAttribute('aria-pressed', selected ? 'false' : 'true');
                    self.controls.setStatus('Fact selection failed: ' + e.message, 'error');
                }
            });
            info.appendChild(row);
        });
    }

    _hydrateBackendToggleState(conceptData) {
        this._edgeVisibility = {};
        this._labelVisibility = {};
        var toggles = (conceptData && conceptData.toggles) || {};
        var edges = toggles.edges || {};
        var labels = toggles.labels || {};
        for (var edge in edges) {
            if (edges.hasOwnProperty(edge)) {
                this._edgeVisibility[edge] = Object.assign({
                    all_to_all: false,
                    edge_unknown: false,
                    none_to_none: false,
                    transitive: false,
                }, edges[edge]);
            }
        }
        for (var label in labels) {
            if (labels.hasOwnProperty(label)) {
                this._labelVisibility[label] = Object.assign({
                    node_necessarily: false,
                    node_maybe: false,
                    node_necessarily_not: false,
                }, labels[label]);
            }
        }
    }

    _toggleChecked(name, displayClass) {
        var base = name.split('(')[0];
        var vis = this._edgeVisibility[name] || this._edgeVisibility[base];
        if (vis && Object.prototype.hasOwnProperty.call(vis, displayClass)) {
            return !!vis[displayClass];
        }
        var labelKeyMap = {
            all_to_all: 'node_necessarily',
            edge_unknown: 'node_maybe',
            none_to_none: 'node_necessarily_not',
        };
        var labelKey = labelKeyMap[displayClass];
        var labelVis = this._labelVisibility[name] || this._labelVisibility[base];
        if (labelKey && labelVis && Object.prototype.hasOwnProperty.call(labelVis, labelKey)) {
            return !!labelVis[labelKey];
        }
        return false;
    }

    /**
     * Handle edge visibility toggle change.
     */
    async onEdgeToggle(edgeName, displayClass, checked) {
        // Track visibility state client-side (matches Python edge/node_label display_checkboxes)
        // Python maps checkbox columns to keys:
        //   For edges:  + → all_to_all, ? → edge_unknown, - → none_to_none, T → transitive
        //   For labels: + → node_necessarily, ? → node_maybe, - → node_necessarily_not
        if (!this._edgeVisibility[edgeName]) {
            this._edgeVisibility[edgeName] = {
                all_to_all: false, edge_unknown: false, none_to_none: false, transitive: false
            };
        }
        this._edgeVisibility[edgeName][displayClass] = checked;

        // Also track as a node label (Python does both: set_checkbox sets BOTH dicts).
        // Store under the bare name (e.g. "semaphore") since _applyNodeLabels
        // looks up by bare node_label names, not the parameterized relation
        // names like "semaphore(X)" that the checkboxes display.
        var labelKeyMap = { 'all_to_all': 'node_necessarily', 'edge_unknown': 'node_maybe', 'none_to_none': 'node_necessarily_not' };
        var labelKey = labelKeyMap[displayClass];
        if (labelKey) {
            var bareName = edgeName.split('(')[0];
            if (!this._labelVisibility[bareName]) {
                this._labelVisibility[bareName] = { node_necessarily: false, node_maybe: false, node_necessarily_not: false };
            }
            this._labelVisibility[bareName][labelKey] = checked;
        }

        // Apply visibility to edges and node labels
        this._applyEdgeVisibility();
        this._applyNodeLabels();

        // Also inform server for persistence
        try {
            await this.api.setToggles({
                edge: edgeName,
                display_class: displayClass,
                value: checked
            });
            await this.refreshConceptGraph();
        } catch (e) {
            console.error('Toggle error:', e);
        }
    }

    /**
     * Apply edge visibility based on checkbox state.
     * An edge is shown if its display class checkbox is checked.
     * Matches Python cy_render.py line 346:
     *   if widget.edge_display_checkboxes[edge][classes[0]].value is False: skip
     */
    _applyEdgeVisibility(conceptGraph) {
        var graph = conceptGraph || this.conceptGraph;
        if (!graph || !graph.cy) return;
        var self = this;
        graph.cy.edges().forEach(function(edge) {
            // Try multiple keys to match: the edge obj, label, and
            // formatted versions like "link(X,Y)" that checkboxes use.
            var obj = edge.data('obj') || '';
            var label = edge.data('label') || '';
            var vis = self._findEdgeVisibility(obj, label);
            if (!vis) {
                // No checkbox state → hide by default (matching Python)
                edge.style('display', 'none');
                return;
            }
            // Check if ANY of the edge's classes has its checkbox checked.
            // Cytoscape classes() returns an array.
            var classList = edge.classes();
            var show = false;
            for (var i = 0; i < classList.length; i++) {
                if (vis[classList[i]]) {
                    show = true;
                    break;
                }
            }
            edge.style('display', show ? 'element' : 'none');
        });
    }

    /**
     * Find edge visibility entry. Checkboxes use display names like "link(X,Y)"
     * while edge data uses bare names like "link". Try both.
     */
    _findEdgeVisibility(obj, label) {
        var ev = this._edgeVisibility;
        // Direct match on obj or label
        if (ev[obj]) return ev[obj];
        if (ev[label]) return ev[label];
        // Checkbox names may include params: "link(X,Y)" — try matching
        // by prefix before the "("
        for (var key in ev) {
            if (ev.hasOwnProperty(key)) {
                var base = key.split('(')[0];
                if (base === obj || base === label) {
                    return ev[key];
                }
            }
        }
        return null;
    }

    /**
     * Apply node label text based on checkbox state.
     * Matches Python cy_render.py render_concept_graph lines 127-146:
     *   For each sort node, check each node_label. If the label's checkbox
     *   is checked, add the label text (with prefix) to the node's display.
     *   Prefix: + → plain, ? → "?suffix", - → "¬prefix"
     */
    /**
     * Apply node label text based on checkbox state and abstract value.
     * Matches Python cy_render.py render_concept_graph lines 127-146 EXACTLY:
     *
     * For each sort node and each node_label:
     *   1. Check if label's sort matches this node's sort (label_sorts map)
     *   2. Determine k from abstract_value:
     *      - if a['node_label|node_necessarily|node|label'] → k = 'node_necessarily'
     *      - elif a['node_label|node_necessarily_not|node|label'] → k = 'node_necessarily_not'
     *      - else → k = 'node_maybe'
     *   3. Check if checkbox[label][k] is checked → if not, skip
     *   4. Display with prefix: '' for +, '?' for ?, '¬' for -
     */
    _applyNodeLabels() {
        if (!this.conceptGraph || !this.conceptGraph.cy) return;
        if (!this._lastConceptData) return;

        var labelPrefixes = {
            'node_necessarily': '',
            'node_maybe': '?',
            'node_necessarily_not': '\\u00AC'
        };

        var nodeLabels = this._lastConceptData.node_labels || [];
        var labelSorts = this._lastConceptData.label_sorts || {};
        var abstractValue = this._lastConceptData.abstract_value || {};

        var self = this;
        this.conceptGraph.cy.nodes().forEach(function (node) {
            var nodeID = node.data('obj') || '';
            if (!nodeID) return;
            var sortName = node.data('cluster') || node.data('sort') || nodeID;
            var topLabel = node.data('display_label') || sortName;

            var labelParts = [topLabel];

            for (var i = 0; i < nodeLabels.length; i++) {
                var labelName = nodeLabels[i];
                var baseLabelName = labelName.split('(')[0];

                // Step 1: Check if this label belongs to this sort node.
                // Python: only adds label to nodes whose sort matches the label's sort.
                var labelSort = labelSorts[labelName] || labelSorts[baseLabelName];
                if (labelSort && labelSort !== sortName) continue;

                // Step 2: Determine k from abstract_value (Z3 result).
                // Python cy_render.py lines 132-137.
                var k;
                var necKey = 'node_label|node_necessarily|' + nodeID + '|' + baseLabelName;
                var necNotKey = 'node_label|node_necessarily_not|' + nodeID + '|' + baseLabelName;
                if (abstractValue[necKey]) {
                    k = 'node_necessarily';
                } else if (abstractValue[necNotKey]) {
                    k = 'node_necessarily_not';
                } else {
                    k = 'node_maybe';
                }

                // Step 3: Check if the checkbox for this k is checked.
                // Python: widget.node_label_display_checkboxes[label_name][k].value
                var vis = self._labelVisibility[labelName] || self._labelVisibility[baseLabelName];
                if (!vis || !vis[k]) continue;

                // Step 4: Display with prefix.
                var prefix = labelPrefixes[k];
                var displayLabelName = self._displayConceptName(baseLabelName);
                if (prefix === '?') {
                    labelParts.push(prefix + displayLabelName);
                } else {
                    labelParts.push(prefix + displayLabelName);
                }
            }

            var newLabel = labelParts.join('\\n');
            if (node.data('label') !== newLabel) {
                node.data('label', newLabel);
                var lines = labelParts.length;
                var h = Math.max(50, 30 + lines * 20);
                node.data('height', h);
            }
        });
    }

    _displayConceptName(name) {
        if (typeof name === 'string' && name.charAt(0) === '=') {
            var body = name.slice(1);
            var idx = body.lastIndexOf(':');
            if (idx > 0) {
                return '=' + body.slice(0, idx);
            }
        }
        return name;
    }

    /**
     * Update the state label to show which ARG node is selected.
     */
    updateStateLabel(nodeId) {
        if (window.__ivyVueBridge && typeof window.__ivyVueBridge.updateStateLabel === 'function') {
            window.__ivyVueBridge.updateStateLabel(nodeId);
            return;
        }
        var label = document.getElementById('state-label');
        if (label) {
            label.textContent = 'State: ' + (nodeId != null ? nodeId : '—');
        }
    }

    /**
     * Set up keyboard shortcuts.
     */
    setupKeyboardShortcuts() {
        if (window.__ivyVueBridge &&
            typeof window.__ivyVueBridge.globalInteractionsHandled === 'function' &&
            window.__ivyVueBridge.globalInteractionsHandled()) {
            return;
        }
        var self = this;
        document.addEventListener('keydown', function (e) {
            // Ctrl+Z or Cmd+Z: Undo
            if ((e.ctrlKey || e.metaKey) && e.key === 'z') {
                e.preventDefault();
                self.doUndo();
            }
            // Escape: hide context menu
            if (e.key === 'Escape') {
                self.controls.hideContextMenu();
            }
        });
    }

    // ================================================================
    // ARG Graph Interactions
    // ================================================================

    /**
     * Handle left-click on an ARG node: load its concept graph.
     */
    async onArgNodeClick(nodeData, sheetId) {
        sheetId = sheetId || this.activeSheetId || 'sheet-1';
        var sheet = this.sheets && this.sheets[sheetId];
        var argGraph = (sheet && sheet.argGraph) || this.argGraph;
        var conceptGraph = (sheet && sheet.conceptGraph) || this.conceptGraph;
        if (sheetId !== this.activeSheetId && this.sheets && this.sheets[sheetId]) {
            this.switchSheet(sheetId);
            sheet = this.sheets[sheetId];
            argGraph = sheet.argGraph;
            conceptGraph = sheet.conceptGraph;
        }
        this.selectedArgNode = nodeData.id;
        if (sheet) {
            sheet.selectedArgNode = nodeData.id;
        }
        if (window.__ivyVueBridge && typeof window.__ivyVueBridge.selectArgNode === 'function') {
            window.__ivyVueBridge.selectArgNode(sheetId, nodeData.id);
        }
        if (argGraph && typeof argGraph.highlightNode === 'function') {
            argGraph.highlightNode(nodeData.id);
        }
        this.controls.showInfo(nodeData.short_info, nodeData.long_info);
        this.updateStateLabel(nodeData.label || nodeData.id);
        if (sheet && sheet.visualOnly) {
            this.controls.setStatus(this.visualOnlyMessage('analysis'), 'warning');
            return;
        }
        this.controls.setStatus('Loading concept graph for state ' + (nodeData.label || nodeData.id) + '...');

        try {
            var result = await this.api.getConceptGraph(nodeData.obj || nodeData.id, sheetId);
            if (result && result.elements) {
                conceptGraph.update(result.elements, result.positions);
            }
            // Build toggles if toggle info is provided
            if (result && result.edge_names) {
                this.controls.buildEdgeToggles(result.edge_names, this.onEdgeToggleChange.bind(this));
            }
            if (result && result.label_names) {
                this.controls.buildLabelToggles(result.label_names, this.onLabelToggleChange.bind(this));
            }
            this.controls.setStatus('Viewing state ' + (nodeData.label || nodeData.id));
        } catch (e) {
            this.controls.setStatus('Error loading concept graph: ' + e.message, 'error');
            console.error('Concept graph load error:', e);
        }
    }

    /**
     * Handle right-click on an ARG node: show context menu.
     * Actions: view state, execute action, mark, cover, safety check.
     */
    onArgNodeRightClick(nodeData, pos, sheetId) {
        var self = this;
        sheetId = sheetId || this.activeSheetId || 'sheet-1';
        var sheet = this.sheets && this.sheets[sheetId];
        if (sheet && sheet.visualOnly) {
            this.controls.setStatus(this.visualOnlyMessage('analysis'), 'warning');
            return;
        }
        var actions = [
            { header: 'State ' + (nodeData.label || nodeData.id) },
            {
                name: 'View State',
                id: 'view_state',
                callback: function () { self.onArgNodeClick(nodeData, sheetId); },
            },
            { separator: true },
        ];

        // Add server-provided actions if available
        if (nodeData.actions && Array.isArray(nodeData.actions)) {
            for (var i = 0; i < nodeData.actions.length; i++) {
                var act = nodeData.actions[i];
                var actionLabel = act.label || act[0] || act.name || '';
                var actionID = act.action || act.id || act[0] || '';
                if (actionLabel === '---') {
                    actions.push({ separator: true });
                    continue;
                }
                if (!actionID) {
                    actions.push({ header: actionLabel });
                    continue;
                }
                (function (action) {
                    actions.push({
                        name: action.label || action[0] || action.name,
                        id: action.action || action.id || action[0],
                        callback: function () {
                            self.executeArgNodeAction(nodeData, action, sheetId);
                        },
                    });
                })(act);
            }
        } else {
            // Default ARG node actions — matches Python ivy_ui.py node_commands()
            var defaultActions = [
                { name: 'Check safety', id: 'check_safety' },
                { name: 'Extend', id: 'extend' },
                { name: 'Mark', id: 'mark' },
                { name: 'Cover by marked', id: 'cover' },
                { name: 'Join with marked', id: 'join' },
                { name: 'Try conjecture', id: 'try_conjecture' },
                { name: 'Try remembered goal', id: 'try_remembered' },
                { name: 'Delete', id: 'delete' },
            ];
            for (var j = 0; j < defaultActions.length; j++) {
                (function (act) {
                    actions.push({
                        name: act.name,
                        id: act.id,
                        callback: function () {
                            self.executeArgNodeAction(nodeData, act, sheetId);
                        },
                    });
                })(defaultActions[j]);
            }
        }

        // Offset from the graph container position
        var graphContainer = document.getElementById((sheet && sheet.argGraph && sheet.argGraph.containerId) || 'arg-graph');
        var rect = graphContainer.getBoundingClientRect();
        this.controls.showContextMenu(rect.left + pos.x, rect.top + pos.y, actions);
    }

    /**
     * Execute an ARG node action via the API.
     */
    async executeArgNodeAction(nodeData, action, sheetId) {
        sheetId = sheetId || this.activeSheetId || 'sheet-1';
        if (this.isVisualOnlySheet(sheetId)) {
            this.controls.setStatus(this.visualOnlyMessage('analysis'), 'warning');
            return null;
        }
        var actionName = action.action || action.id || action[0] || action.name;
        this.controls.setStatus('Executing: ' + actionName + '...');
        try {
            var args = Object.assign({}, action.args || {});
            args.sheet_id = sheetId;
            args = await this.prepareArgNodeActionArgs(nodeData, actionName, args, sheetId);
            if (args === null) {
                this.controls.setStatus('Action cancelled: ' + actionName, 'warning');
                return;
            }
            var result = await this.api.argNodeAction(nodeData.obj || nodeData.id, actionName, args);
            var sheet = this.sheets && this.sheets[sheetId];
            var argGraph = (sheet && sheet.argGraph) || this.argGraph;
            var conceptGraph = (sheet && sheet.conceptGraph) || this.conceptGraph;
            if (result && result.arg) {
                argGraph.update(result.arg.elements, result.arg.positions);
            }
            if (result && result.concept) {
                conceptGraph.update(result.concept.elements, result.concept.positions);
            }
            this.controls.setStatus('Action complete: ' + actionName, 'success');
        } catch (e) {
            this.controls.setStatus('Action failed: ' + e.message, 'error');
            console.error('ARG action error:', e);
        }
    }

    async prepareArgNodeActionArgs(nodeData, actionName, args, sheetId) {
        if (actionName === 'try_conjecture' && !args.conjecture) {
            var conjChoices = await this.api.argNodeAction(nodeData.obj || nodeData.id, 'try_conjecture_choices', { sheet_id: sheetId });
            var selectedConj = await this.listboxDialog(
                'Try conjecture',
                'Choose a conjecture to prove:',
                conjChoices.choices || [],
                { cancel: true }
            );
            if (selectedConj == null) return null;
            args.conjecture = selectedConj;
        } else if (actionName === 'try_remembered' && !args.goal) {
            var goalChoices = await this.api.argNodeAction(nodeData.obj || nodeData.id, 'try_remembered_choices', { sheet_id: sheetId });
            var selectedGoal = await this.listboxDialog(
                'Try remembered goal',
                'Choose a remembered goal:',
                goalChoices.choices || [],
                { cancel: true }
            );
            if (selectedGoal == null) return null;
            args.goal = selectedGoal;
        }
        return args;
    }

    /**
     * Handle right-click on an ARG edge: show context menu.
     * Matches Python ivy_ui.py get_edge_actions: Dismiss, Recalculate, Step in, View Source.
     */
    onArgEdgeRightClick(edgeData, pos, sheetId) {
        var self = this;
        sheetId = sheetId || this.activeSheetId || 'sheet-1';
        var sheet = this.sheets && this.sheets[sheetId];
        var label = edgeData.label || edgeData.obj || '';
        var actions = [
            { header: 'Transition: ' + label },
            {
                name: 'Dismiss',
                id: 'dismiss',
                callback: function () { self.controls.hideContextMenu(); },
            },
            {
                name: 'Recalculate',
                id: 'recalculate_edge',
                callback: function () { self.executeArgEdgeAction(edgeData, 'recalculate', sheetId); },
            },
            {
                name: 'Step in',
                id: 'decompose_edge',
                callback: function () { self.executeArgEdgeAction(edgeData, 'decompose', sheetId); },
            },
            {
                name: 'View Source',
                id: 'view_source_edge',
                callback: function () { self.executeArgEdgeAction(edgeData, 'view_source', sheetId); },
            },
        ];
        var graphContainer = document.getElementById((sheet && sheet.argGraph && sheet.argGraph.containerId) || 'arg-graph');
        var rect = graphContainer.getBoundingClientRect();
        this.controls.showContextMenu(rect.left + pos.x, rect.top + pos.y, actions);
    }

    /**
     * Execute an ARG edge action via the API.
     */
    async executeArgEdgeAction(edgeData, actionName, sheetId) {
        sheetId = sheetId || this.activeSheetId || 'sheet-1';
        this.controls.setStatus('Executing: ' + actionName + '...');
        try {
            var result = await this.api.argNodeAction(
                edgeData.source_obj || edgeData.source || edgeData.obj,
                actionName,
                { target: edgeData.target_obj || edgeData.target, sheet_id: sheetId }
            );
            if (actionName === 'decompose' && result && result.decomposed) {
                // Decompose: open a new tab with the sub-ARG
                var label = 'Step: ' + (edgeData.label || actionName);
                this.openARGSheet(label, result.sub_arg, result.sheet_id);
            }
            if (result && result.arg) {
                var sheet = this.sheets && this.sheets[sheetId];
                var argGraph = (sheet && sheet.argGraph) || this.argGraph;
                argGraph.update(result.arg.elements, result.arg.positions);
            }
            if (result && result.source && actionName === 'view_source') {
                // Show source in the model editor and scroll to the action line.
                // Matches Python ivy_ui.py view_source_edge → browse(filename, lineno).
                this.setEditorContent(result.source);
                if (result.lineno) {
                    this.scrollEditorToLine(result.lineno);
                }
                this.controls.showInfo(
                    'Source: ' + (result.file || '') + (result.lineno ? ' line ' + result.lineno : ''),
                    ''
                );
            }
            this.controls.setStatus('Done: ' + actionName, 'success');
        } catch (e) {
            this.controls.setStatus('Edge action failed: ' + e.message, 'error');
            console.error('ARG edge action error:', e);
        }
    }

    // ================================================================
    // Concept Graph Interactions
    // ================================================================

    /**
     * Handle right-click on a concept node: show context menu.
     * Actions: split by X (for each possible split), suppose empty, remove, materialize.
     */
    onConceptNodeRightClick(nodeData, pos) {
        var self = this;
        var actions = [
            { header: nodeData.label || nodeData.obj || nodeData.id },
        ];

        // Server-provided actions (split by X, remove, empty, materialize, add projection)
        if (nodeData.actions && Array.isArray(nodeData.actions)) {
            for (var i = 0; i < nodeData.actions.length; i++) {
                var act = nodeData.actions[i];
                var label = act.label || act[0] || act.name || '';
                var id = act.action || act.id || act[0] || '';
                if (label === '---') {
                    actions.push({ separator: true });
                    continue;
                }
                if (!id) {
                    actions.push({ header: label });
                    continue;
                }
                (function (action) {
                    var actionName = action.label || action[0] || action.name;
                    actions.push({
                        name: actionName,
                        id: action.action || action.id || action[0] || actionName,
                        callback: function () {
                            self.executeConceptNodeAction(nodeData, action);
                        },
                    });
                })(act);
            }
        } else {
            // Default concept node actions — matches Python tk_graph_ui.py
            var conceptId = nodeData.obj || nodeData.id;
            var isSelected = false;
            if (self.conceptGraph && self.conceptGraph.cy) {
                var nd = self.conceptGraph.cy.nodes().filter(function (n) {
                    return n.data('obj') === conceptId || n.id() === conceptId;
                });
                isSelected = nd.length > 0 && nd.hasClass('selected_node');
            }
            actions.push({
                name: isSelected ? 'Unselect' : 'Select',
                id: 'select',
                callback: function () {
                    self.selectConceptNode(conceptId);
                },
            });
            actions.push({
                name: 'Suppose Empty',
                id: 'suppose_empty',
                callback: function () {
                    self.supposeEmpty(nodeData.obj || nodeData.id);
                },
            });
            actions.push({
                name: 'Materialize',
                id: 'materialize',
                callback: function () {
                    self.materializeNode(nodeData.obj || nodeData.id);
                },
            });
            actions.push({
                name: 'Materialize edge',
                id: 'materialize_from_selected',
                callback: function () {
                    self.materializeEdgeFromSelected(nodeData.obj || nodeData.id);
                },
            });
            actions.push({
                name: 'Splatter',
                id: 'splatter',
                callback: function () {
                    self.splatterNode(nodeData.obj || nodeData.id);
                },
            });
            actions.push({ separator: true });
            actions.push({
                name: 'Remove',
                id: 'remove',
                callback: function () {
                    self.removeConcept(nodeData.obj || nodeData.id);
                },
            });
        }

        var graphContainer = document.getElementById('concept-graph');
        var rect = graphContainer.getBoundingClientRect();
        this.controls.showContextMenu(rect.left + pos.x, rect.top + pos.y, actions);
    }

    /**
     * Handle right-click on a concept edge: show context menu.
     * Actions: remove, materialize +, materialize -.
     */
    onConceptEdgeRightClick(edgeData, pos) {
        var self = this;
        var conceptId = edgeData.obj || edgeData.id;
        var actions = [
            { header: edgeData.label || conceptId },
        ];

        if (edgeData.actions && Array.isArray(edgeData.actions)) {
            for (var i = 0; i < edgeData.actions.length; i++) {
                var act = edgeData.actions[i];
                (function (action) {
                    var actionName = action[0] || action.name;
                    actions.push({
                        name: actionName,
                        id: actionName,
                        callback: function () {
                            self.executeConceptEdgeAction(edgeData, action);
                        },
                    });
                })(act);
            }
        } else {
            actions.push({
                name: 'Remove',
                id: 'remove',
                callback: function () { self.removeConcept(conceptId); },
            });
            actions.push({
                name: 'Materialize +',
                id: 'materialize_pos',
                callback: function () { self.materializeEdge(edgeData, true); },
            });
            actions.push({
                name: 'Materialize \\u2013',
                id: 'materialize_neg',
                callback: function () { self.materializeEdge(edgeData, false); },
            });
            actions.push({ separator: true });
            actions.push({
                name: 'Suppose Empty',
                id: 'empty_edge',
                callback: function () { self.supposeEmpty(conceptId); },
            });
            actions.push({
                name: 'Dematerialize',
                id: 'dematerialize',
                callback: function () {
                    self.executeConceptEdgeAction(edgeData, { id: 'dematerialize' });
                },
            });
        }

        var graphContainer = document.getElementById('concept-graph');
        var rect = graphContainer.getBoundingClientRect();
        this.controls.showContextMenu(rect.left + pos.x, rect.top + pos.y, actions);
    }

    /**
     * Execute a concept node action dispatched from the context menu.
     */
    async executeConceptNodeAction(nodeData, action) {
        var actionID = action.action || action.id || action[0] || action.name || '';
        var actionName = actionID.toLowerCase();
        var actionArgs = action.args || {};

        if (actionName === 'remove') {
            return this.removeConcept(nodeData.obj || nodeData.id);
        }
        if (actionName === 'suppose empty') {
            return this.supposeEmpty(nodeData.obj || nodeData.id);
        }
        if (actionName === 'materialize') {
            return this.materializeNode(nodeData.obj || nodeData.id);
        }
        if (actionName === 'materialize edge' || actionName === 'materialize_from_selected') {
            return this.materializeEdgeFromSelected(nodeData.obj || nodeData.id);
        }
        if (actionName.indexOf('split by ') === 0) {
            var splitBy = actionName.substring('split by '.length);
            return this.splitConcept(nodeData.obj || nodeData.id, splitBy);
        }
        if (actionName.indexOf('add ') === 0) {
            var projName = actionName.substring('add '.length);
            return this.addProjection(projName, nodeData.obj || nodeData.id);
        }
        if (actionName === 'add_projection') {
            return this.addProjection(actionArgs.name, actionArgs.concept || actionArgs.name);
        }

        // Fallback: generic action execution
        this.controls.setStatus('Executing: ' + actionName + '...');
        try {
            var result = await this.api.executeAction(actionName, {
                concept: nodeData.obj || nodeData.id,
            });
            await this.refreshConceptGraph();
            this.controls.setStatus('Done: ' + actionName, 'success');
        } catch (e) {
            this.controls.setStatus('Error: ' + e.message, 'error');
        }
    }

    /**
     * Execute a concept edge action dispatched from the context menu.
     */
    async executeConceptEdgeAction(edgeData, action) {
        var actionName = (action[0] || action.name || '').toLowerCase();
        var conceptId = edgeData.obj || edgeData.id;

        if (actionName === 'remove') {
            return this.removeConcept(conceptId);
        }
        if (actionName === 'materialize +') {
            return this.materializeEdge(edgeData, true);
        }
        if (actionName === 'materialize -' || actionName === 'materialize \\u2013') {
            return this.materializeEdge(edgeData, false);
        }
        if (actionName === 'dematerialize') {
            return this.materializeEdge(edgeData, false);
        }

        // Fallback
        this.controls.setStatus('Executing: ' + actionName + '...');
        try {
            await this.api.executeAction(actionName, { concept: conceptId });
            await this.refreshConceptGraph();
            this.controls.setStatus('Done: ' + actionName, 'success');
        } catch (e) {
            this.controls.setStatus('Error: ' + e.message, 'error');
        }
    }

    // ================================================================
    // Concept Domain Operations
    // ================================================================

    async splitConcept(concept, splitBy) {
        this.controls.setStatus('Splitting ' + concept + ' by ' + splitBy + '...');
        try {
            await this.api.splitConcept(concept, splitBy);
            await this.refreshConceptGraph();
            this.controls.setStatus('Split complete', 'success');
        } catch (e) {
            this.controls.setStatus('Split failed: ' + e.message, 'error');
        }
    }

    async supposeEmpty(concept) {
        this.controls.setStatus('Supposing ' + concept + ' is empty...');
        try {
            await this.api.supposeEmpty(concept);
            await this.refreshConceptGraph();
            this.controls.setStatus('Suppose empty applied', 'success');
        } catch (e) {
            this.controls.setStatus('Suppose empty failed: ' + e.message, 'error');
        }
    }

    async removeConcept(concept) {
        this.controls.setStatus('Removing concept...');
        try {
            await this.api.removeConcept(concept);
            await this.refreshConceptGraph();
            this.controls.setStatus('Concept removed', 'success');
        } catch (e) {
            this.controls.setStatus('Remove failed: ' + e.message, 'error');
        }
    }

    async materializeNode(concept) {
        this.controls.setStatus('Materializing node...');
        try {
            await this.api.materializeNode(concept);
            await this.refreshConceptGraph();
            this.controls.setStatus('Node materialized', 'success');
        } catch (e) {
            this.controls.setStatus('Materialize failed: ' + e.message, 'error');
        }
    }

    async materializeEdge(edgeOrConcept, positive) {
        var dir = positive ? '+' : '\\u2013';
        var relation = edgeOrConcept;
        var source = '';
        var target = '';
        if (edgeOrConcept && typeof edgeOrConcept === 'object') {
            relation = edgeOrConcept.obj || edgeOrConcept.label || edgeOrConcept.id;
            source = edgeOrConcept.source_obj || edgeOrConcept.source || '';
            target = edgeOrConcept.target_obj || edgeOrConcept.target || '';
        }
        this.controls.setStatus('Materializing edge (' + dir + ')...');
        try {
            await this.api.materializeEdge(relation, source, target, positive);
            await this.refreshConceptGraph();
            this.controls.setStatus('Edge materialized (' + dir + ')', 'success');
        } catch (e) {
            this.controls.setStatus('Materialize failed: ' + e.message, 'error');
        }
    }

    async addProjection(name, concept) {
        this.controls.setStatus('Adding projection ' + name + '...');
        try {
            await this.api.addProjection(name, concept);
            await this.refreshConceptGraph();
            this.controls.setStatus('Projection added', 'success');
        } catch (e) {
            this.controls.setStatus('Add projection failed: ' + e.message, 'error');
        }
    }

    /**
     * Select/mark a concept node for edge materialization.
     * Matches Python tk_graph_ui.py select action.
     */
    selectConceptNode(conceptId) {
        this.selectedConceptNode = conceptId;
        // Toggle selected_node class (same visual as left-click)
        if (this.conceptGraph && this.conceptGraph.cy) {
            var node = this.conceptGraph.cy.nodes().filter(function (n) {
                return n.data('obj') === conceptId || n.id() === conceptId;
            });
            if (node.length > 0) {
                if (node.hasClass('selected_node')) {
                    node.removeClass('selected_node');
                    this.controls.setStatus('Deselected: ' + conceptId);
                } else {
                    node.addClass('selected_node');
                    this.controls.setStatus('Selected: ' + conceptId);
                }
            }
        }
    }

    async materializeEdgeFromSelected(targetConceptId) {
        var sourceConceptId = this.selectedConceptNode;
        if (!sourceConceptId) {
            this.controls.setStatus('Select a source node first', 'warning');
            return;
        }
        var data = this._lastConceptData || {};
        var edgeSorts = data.edge_sorts || {};
        var relations = Object.keys(edgeSorts).filter(function (rel) {
            var sorts = edgeSorts[rel] || [];
            return sorts.length >= 2 && sorts[0] === sourceConceptId && sorts[1] === targetConceptId;
        });
        if (relations.length === 0 && Array.isArray(data.edges)) {
            relations = data.edges.slice();
        }
        if (relations.length === 0) {
            this.controls.setStatus('No matching binary relations', 'warning');
            return;
        }
        var selected = await this.listboxDialog(
            'Materialize edge',
            'Materialize this relation from selected node:',
            relations.map(function (rel) { return { label: rel, value: rel }; }),
            { cancel: true }
        );
        if (selected == null) {
            this.controls.setStatus('Materialize edge cancelled', 'warning');
            return;
        }
        this.controls.setStatus('Materializing edge ' + selected + '...');
        try {
            await this.api.materializeEdge(selected, sourceConceptId, targetConceptId, true);
            await this.refreshConceptGraph();
            this.controls.setStatus('Edge materialized (+)', 'success');
        } catch (e) {
            this.controls.setStatus('Materialize edge failed: ' + e.message, 'error');
        }
    }

    /**
     * Splatter a concept node — materialize all universe elements of its sort.
     * Matches Python tk_graph_ui.py splatter action.
     */
    async splatterNode(conceptId) {
        this.controls.setStatus('Splattering ' + conceptId + '...');
        try {
            await this.api.executeAction('splatter', { concept: conceptId });
            await this.refreshConceptGraph();
            this.controls.setStatus('Splattered: ' + conceptId, 'success');
        } catch (e) {
            this.controls.setStatus('Splatter failed: ' + e.message, 'error');
        }
    }

    // ================================================================
    // Toggle Handlers
    // ================================================================

    /**
     * Called when an edge visibility toggle checkbox changes.
     */
    async onEdgeToggleChange(edgeName, className, checked) {
        try {
            var toggles = {
                edge: edgeName,
                display_class: className,
                value: checked,
            };
            await this.api.setToggles(toggles);
            await this.refreshConceptGraph();
        } catch (e) {
            console.error('Toggle update error:', e);
        }
    }

    /**
     * Called when a label visibility toggle checkbox changes.
     */
    async onLabelToggleChange(labelName, className, checked) {
        try {
            var toggles = {
                label: labelName,
                display_class: className,
                value: checked,
            };
            await this.api.setToggles(toggles);
            await this.refreshConceptGraph();
        } catch (e) {
            console.error('Toggle update error:', e);
        }
    }

    // ================================================================
    // Top-Level Actions
    // ================================================================

    /**
     * Load an .ivy file.
     */
    async loadFile(file) {
        this.controls.showLoading('Loading ' + file.name + '...');
        this.controls.setStatus('Loading file: ' + file.name + '...');
        try {
            // Read file content for persistence before uploading
            var self = this;
            var reader = new FileReader();
            var contentPromise = new Promise(function (resolve) {
                reader.onload = function () { resolve(reader.result); };
                reader.readAsText(file);
            });
            var fileContent = await contentPromise;
            self._persistedFileName = file.name;
            self._persistedFilePath = file.path || file.webkitRelativePath || file.name;
            self._persistedFileContent = fileContent;
            if (self._fileHandle) {
                await IvyPersist.saveFileHandle(self);
            }

            // Populate the model editor with the file content (also marks clean via setEditorContent)
            this.setEditorContent(fileContent);

            var result = await this.api.loadFile(file);
            // Refresh ARG
            var argData = await this.api.getARG();
            if (argData && argData.elements) {
                this.argGraph.update(argData.elements, argData.positions);
            }
            // Refresh concept graph and populate state checkbox pane
            var conceptData = await this.api.getConceptGraph();
            if (conceptData && conceptData.elements) {
                this.conceptGraph.update(conceptData.elements, conceptData.positions);
            }
            this._persistedConceptRelations = conceptData;
            this.populateStateCheckboxes(conceptData);
            // Update state label and file name display
            this.updateStateLabel(0);
            IvyPersist.setFileName(file.name, this._persistedFilePath);
            this.controls.setStatus('Loaded: ' + file.name, 'success');

            // Auto-save after file load
            IvyPersist.save(this);
        } catch (e) {
            this.controls.setStatus('Load failed: ' + e.message, 'error');
            console.error('File load error:', e);
        } finally {
            this.controls.hideLoading();
        }
    }

    /**
     * Save invariant/conjectures to a .ivy file.
     * Matches Python ivy_ui_cti.py save_conjectures:
     * exports conjectures as "invariant [label] formula" lines.
     */
    async saveInvariant() {
        this.controls.setStatus('Saving invariant...');
        try {
            var result = await this.api.executeAction('save_invariant', {});
            var text = (result && result.content) || '';
            var suggestedName = (this._persistedFileName || 'model').replace(/\\.ivy$/, '') + '_invariant.ivy';

            // Use File System Access API to let user choose save location
            if (window.showSaveFilePicker) {
                var handle = await window.showSaveFilePicker({
                    suggestedName: suggestedName,
                    types: [{
                        description: 'Ivy files',
                        accept: { 'text/plain': ['.ivy'] },
                    }],
                });
                var writable = await handle.createWritable();
                await writable.write(text);
                await writable.close();
                this.controls.setStatus('Invariant saved: ' + handle.name, 'success');
            } else {
                // Fallback: browser download
                var blob = new Blob([text], { type: 'text/plain' });
                var url = URL.createObjectURL(blob);
                var a = document.createElement('a');
                a.href = url;
                a.download = suggestedName;
                document.body.appendChild(a);
                a.click();
                document.body.removeChild(a);
                URL.revokeObjectURL(url);
                this.controls.setStatus('Invariant downloaded: ' + suggestedName, 'success');
            }
        } catch (e) {
            if (e.name === 'AbortError') {
                this.controls.setStatus('Save invariant cancelled');
            } else {
                this.controls.setStatus('Save invariant failed: ' + e.message, 'error');
            }
        }
    }

    /**
     * Download current model as a browser download.
     */
    async downloadModel() {
        this.controls.setStatus('Downloading...');
        try {
            var content = this._editorContent();
            if (!content) {
                this.controls.setStatus('No model loaded to download', 'error');
                return;
            }
            var filename = this._persistedFileName || 'model.ivy';
            this.downloadTextFile(filename, content, 'text/plain');
            this.controls.setStatus('Downloaded: ' + filename, 'success');
        } catch (e) {
            this.controls.setStatus('Download failed: ' + e.message, 'error');
        }
    }

    downloadTextFile(filename, content, mimeType) {
        var blob = new Blob([content || ''], { type: mimeType || 'text/plain' });
        var url = URL.createObjectURL(blob);
        var a = document.createElement('a');
        a.href = url;
        a.download = filename || 'download.txt';
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        URL.revokeObjectURL(url);
    }

    /**
     * Save as... — uses the File System Access API (showSaveFilePicker)
     * to let the user choose a disk path. Remembers the file handle
     * for subsequent saves.
     */
    async save() {
        var content = this._editorContent();
        var dirty = this._editorDirty();
        var saveProgress = dirty ? this._showSaveProgress('Saving...') : null;
        try {
            await this._restoreFileHandleForCurrentFile();
            if (this._fileHandle) {
                var writableAllowed = await this._ensureFileHandleWritable();
                if (!writableAllowed) {
                    this.controls.setStatus('Save permission denied', 'error');
                    return false;
                }
                var saveDecision = await this._confirmNoExternalChangeBeforeSave(content);
                if (saveDecision === 'skip') {
                    return false;
                }
                if (!dirty && saveDecision !== 'overwrite') {
                    this._updateEditorLabel();
                    return true;
                }
                var writable = await this._fileHandle.createWritable();
                await writable.write(content);
                await writable.close();
                this._persistedFileContent = content;
                this._savedFileContent = content;
                this._updateEditorLabel();
                this.controls.setStatus('Saved: ' + this._persistedFileName, 'success');
                return true;
            } else {
                if (saveProgress && window.showSaveFilePicker) {
                    this._hideSaveProgress();
                    saveProgress = null;
                }
                if (!dirty) {
                    this._updateEditorLabel();
                    return true;
                }
                if (!window.showSaveFilePicker) {
                    return this.downloadModelForUnsupportedSave(content);
                }
                return await this.saveAs({ explainMissingHandle: true });
            }
        } catch (e) {
            this.controls.setStatus('Save failed: ' + e.message, 'error');
            return false;
        } finally {
            if (saveProgress) {
                this._hideSaveProgress();
            }
        }
    }

    downloadModelForUnsupportedSave(content) {
        var filename = this._persistedFileName || 'model.ivy';
        try {
            this.downloadTextFile(filename, content, 'text/plain');
            this._persistedFileContent = content;
            this._savedFileContent = content;
            this._updateEditorLabel();
            IvyPersist.save(this);
            this.controls.setStatus('Downloaded edited copy: ' + filename + '. In Firefox, choose the original file to overwrite.', 'success');
            return true;
        } catch (e) {
            this.controls.setStatus('Download failed: ' + e.message, 'error');
            return false;
        }
    }

    async saveAs(options) {
        options = options || {};
        var content = this._editorContent();
        if (!content) {
            this.controls.setStatus('No model loaded to save', 'error');
            return false;
        }
        var saveProgress = null;
        try {
            if (!window.showSaveFilePicker) {
                return this.downloadModelForUnsupportedSave(content);
            }
            if (options.explainMissingHandle) {
                this.showSaveAsExplanationNotice();
            }
            var handle;
            try {
                handle = await window.showSaveFilePicker({
                    suggestedName: this._persistedFileName || 'model.ivy',
                    types: [{
                        description: 'Ivy files',
                        accept: { 'text/plain': ['.ivy'] },
                    }],
                });
            } finally {
                if (options.explainMissingHandle) {
                    this.hideSaveAsExplanationNotice();
                }
            }
            saveProgress = this._showSaveProgress('Saving...');
            var writable = await handle.createWritable();
            await writable.write(content);
            await writable.close();

            // Remember the handle and path for future saves
            this._fileHandle = handle;
            this._persistedFileName = handle.name;
            this._persistedFilePath = handle.name;
            this._persistedFileContent = content;
            this._savedFileContent = content;
            await IvyPersist.saveFileHandle(this);
            IvyPersist.setFileName(handle.name, this._persistedFilePath);
            this._updateEditorLabel();
            this.controls.setStatus('Saved: ' + handle.name, 'success');
            return true;
        } catch (e) {
            if (e.name === 'AbortError') {
                // User cancelled the dialog
                this.controls.setStatus('Save cancelled');
            } else {
                this.controls.setStatus('Save as failed: ' + e.message, 'error');
            }
            return false;
        } finally {
            if (saveProgress) {
                this._hideSaveProgress();
            }
        }
    }

    // Keep saveSession as an alias for downloadModel (used by ARG panel binding)
    async saveSession() { return this.downloadModel(); }

    graphElementsSnapshot(graph) {
        if (!graph || !graph.cy || typeof graph.cy.json !== 'function') return null;
        var json = graph.cy.json();
        return json ? json.elements : null;
    }

    getEditorKeymap() {
        if (window.__ivyVueBridge && typeof window.__ivyVueBridge.getEditorKeymap === 'function') {
            return window.__ivyVueBridge.getEditorKeymap() || 'sublime';
        }
        var checked = document.querySelector('input[name="keymap"]:checked');
        return checked ? checked.value : 'sublime';
    }

    setEditorKeymap(keymap) {
        var allowed = { sublime: true, emacs: true, vim: true };
        var next = allowed[keymap] ? keymap : 'sublime';
        if (this.cmEditor && typeof this.cmEditor.setOption === 'function') {
            this.cmEditor.setOption('keyMap', next);
        }
        if (window.__ivyVueBridge && typeof window.__ivyVueBridge.setEditorKeymap === 'function') {
            window.__ivyVueBridge.setEditorKeymap(next);
            return;
        }
        var radio = document.querySelector('input[name="keymap"][value="' + next + '"]');
        if (radio) radio.checked = true;
    }

    tabLabelForSheet(sheetId) {
        if (window.__ivyVueBridge && typeof window.__ivyVueBridge.getSheetTabLabel === 'function') {
            var bridgeLabel = window.__ivyVueBridge.getSheetTabLabel(sheetId);
            if (bridgeLabel) return bridgeLabel;
        }
        var tab = this.sheetTab(sheetId);
        var label = tab ? tab.querySelector('span') : null;
        return label ? label.textContent : sheetId;
    }

    getMode() {
        if (window.__ivyVueBridge && typeof window.__ivyVueBridge.getMode === 'function') {
            return window.__ivyVueBridge.getMode() || 'pdr';
        }
        var modeEl = document.getElementById('mode-select');
        return modeEl ? modeEl.value : 'pdr';
    }

    setMode(mode) {
        if (window.__ivyVueBridge && typeof window.__ivyVueBridge.setMode === 'function') {
            window.__ivyVueBridge.setMode(mode);
            return;
        }
        var modeEl = document.getElementById('mode-select');
        if (modeEl && mode) modeEl.value = mode;
    }

    buildAnalysisState() {
        var sheets = [];
        var ids = Object.keys(this.sheets || {});
        ids.sort(function (a, b) {
            if (a === 'sheet-1') return -1;
            if (b === 'sheet-1') return 1;
            return a.localeCompare(b);
        });
        for (var i = 0; i < ids.length; i++) {
            var sheetId = ids[i];
            var sheet = this.sheets[sheetId];
            if (!sheet) continue;
            if (sheet.type === 'events') {
                sheets.push({
                    id: sheetId,
                    type: 'events',
                    label: this.tabLabelForSheet(sheetId),
                    events: sheet.events || [],
                    patterns: sheet.patterns || [],
                    selectedEventAddress: sheet.selectedEventAddress || null,
                });
            } else {
                sheets.push({
                    id: sheetId,
                    type: 'analysis',
                    label: this.tabLabelForSheet(sheetId),
                    selectedArgNode: sheet.selectedArgNode || null,
                    arg: { elements: this.graphElementsSnapshot(sheet.argGraph), positions: null },
                    concept: { elements: this.graphElementsSnapshot(sheet.conceptGraph), positions: null },
                });
            }
        }
        return {
            analysis_state_format: 'ivyweb-json',
            analysis_state_version: 1,
            python_a2g_equivalent: false,
            fileName: this._persistedFileName || '',
            filePath: this._persistedFilePath || this._persistedFileName || '',
            fileContent: this._editorContent ? this._editorContent() : (this._persistedFileContent || ''),
            mode: this.getMode(),
            activeSheetId: this.activeSheetId || 'sheet-1',
            selectedArgNode: this.selectedArgNode || null,
            edgeVisibility: this._edgeVisibility || {},
            labelVisibility: this._labelVisibility || {},
            toggles: IvyPersist._getToggles ? IvyPersist._getToggles() : {},
            sheets: sheets,
        };
    }

    async saveAnalysisState() {
        try {
            var state = this.buildAnalysisState();
            var text = JSON.stringify(state, null, 2) + '\\n';
            var suggestedName = (this._persistedFileName || 'ivy_analysis').replace(/\\.ivy$/, '') + '.ivyweb.json';
            if (window.showSaveFilePicker) {
                var handle = await window.showSaveFilePicker({
                    suggestedName: suggestedName,
                    types: [{ description: 'IvyWeb analysis state', accept: { 'application/json': ['.json'] } }],
                });
                var writable = await handle.createWritable();
                await writable.write(text);
                await writable.close();
                this.controls.setStatus('Analysis state saved: ' + handle.name, 'success');
            } else {
                this.downloadTextFile(suggestedName, text, 'application/json');
                this.controls.setStatus('Analysis state downloaded: ' + suggestedName, 'success');
            }
            return state;
        } catch (e) {
            if (e.name === 'AbortError') {
                this.controls.setStatus('Save analysis state cancelled');
            } else {
                this.controls.setStatus('Save analysis state failed: ' + e.message, 'error');
            }
            return null;
        }
    }

    async loadAnalysisStateFile(file) {
        if (!file) return false;
        if (typeof file.size === 'number' && file.size > this.analysisStateLimits().maxFileBytes) {
            throw new Error('analysis state file too large');
        }
        var text = await this.readFileText(file);
        return this.loadAnalysisStateObject(JSON.parse(text));
    }

    async loadAnalysisStateObject(state) {
        this.validateAnalysisStateObject(state);
        if (!state || state.analysis_state_format !== 'ivyweb-json') {
            throw new Error('unsupported analysis state format');
        }
        this._persistedFileName = state.fileName || '';
        this._persistedFilePath = state.filePath || state.fileName || '';
        this._persistedFileContent = state.fileContent || '';
        this._savedFileContent = this._persistedFileContent;
        this._edgeVisibility = state.edgeVisibility || {};
        this._labelVisibility = state.labelVisibility || {};
        this.selectedArgNode = state.selectedArgNode || null;

        if (this.setEditorContent) {
            this.setEditorContent(this._persistedFileContent);
        }
        if (state.mode) this.setMode(state.mode);
        if (this.api && this.api.reloadContent && this._persistedFileContent) {
            await this.api.reloadContent(this._persistedFileContent, this._persistedFileName || 'restored.ivy');
        }

        this.removeAnalysisStateExtraSheets();
        var sheets = state.sheets || [];
        for (var i = 0; i < sheets.length; i++) {
            var sheet = sheets[i];
            if (sheet.type === 'events') {
                this.openEventTraceSheet(sheet.label || 'Events', {
                    sheet_id: sheet.id,
                    events: sheet.events || [],
                    patterns: sheet.patterns || [],
                    selected_address: sheet.selectedEventAddress || null,
                }, sheet.id);
                this.setVisualOnlySheet(sheet.id, true);
                continue;
            }
            if (sheet.id === 'sheet-1') {
                if (sheet.arg && sheet.arg.elements && this.argGraph) {
                    this.argGraph.update(sheet.arg.elements, sheet.arg.positions || undefined);
                }
                if (sheet.concept && sheet.concept.elements && this.conceptGraph) {
                    this.conceptGraph.update(sheet.concept.elements, sheet.concept.positions || undefined);
                }
                if (this.sheets && this.sheets['sheet-1']) {
                    this.sheets['sheet-1'].selectedArgNode = sheet.selectedArgNode || null;
                    this.sheets['sheet-1'].visualOnly = true;
                }
            } else if (sheet.arg && sheet.arg.elements) {
                this.openARGSheet(sheet.label || sheet.id, sheet.arg, sheet.id);
                var opened = this.sheets && this.sheets[sheet.id];
                if (opened && opened.conceptGraph && sheet.concept && sheet.concept.elements) {
                    opened.conceptGraph.update(sheet.concept.elements, sheet.concept.positions || undefined);
                }
                if (opened) {
                    opened.selectedArgNode = sheet.selectedArgNode || null;
                    opened.visualOnly = true;
                }
            }
        }

        if (state.toggles) {
            IvyPersist._setToggles(state.toggles);
        }
        if (state.activeSheetId && this.sheetExists(state.activeSheetId)) {
            this.switchSheet(state.activeSheetId);
        } else {
            this.switchSheet('sheet-1');
        }
        IvyPersist.setFileName(this._persistedFileName, this._persistedFilePath);
        this.controls.setStatus('Visual analysis state loaded: ' + (this._persistedFileName || 'state'), 'warning');
        return true;
    }

    analysisStateLimits() {
        return {
            maxFileBytes: 25 * 1024 * 1024,
            maxSheets: 100,
            maxGraphElements: 50000,
            maxEvents: 100000,
            maxEventDepth: 200,
        };
    }

    validateAnalysisStateObject(state) {
        if (!state || state.analysis_state_format !== 'ivyweb-json') {
            throw new Error('unsupported analysis state format');
        }
        var limits = this.analysisStateLimits();
        var sheets = state.sheets || [];
        if (!Array.isArray(sheets)) {
            throw new Error('analysis state sheets must be an array');
        }
        if (sheets.length > limits.maxSheets) {
            throw new Error('too many sheets in analysis state');
        }
        if (state.activeSheetId && !this.isValidSheetId(state.activeSheetId)) {
            throw new Error('invalid active sheet id: ' + state.activeSheetId);
        }
        var seen = {};
        for (var i = 0; i < sheets.length; i++) {
            this.validateAnalysisStateSheet(sheets[i], seen, limits);
        }
    }

    validateAnalysisStateSheet(sheet, seen, limits) {
        if (!sheet || typeof sheet !== 'object') {
            throw new Error('invalid sheet entry');
        }
        if (!this.isValidSheetId(sheet.id)) {
            throw new Error('invalid sheet id: ' + sheet.id);
        }
        if (seen[sheet.id]) {
            throw new Error('duplicate sheet id: ' + sheet.id);
        }
        seen[sheet.id] = true;
        if (sheet.type !== 'analysis' && sheet.type !== 'events') {
            throw new Error('invalid sheet type: ' + sheet.type);
        }
        if (sheet.type === 'events') {
            if (sheet.events != null && !Array.isArray(sheet.events)) {
                throw new Error('event sheet events must be an array');
            }
            if (sheet.patterns != null && !Array.isArray(sheet.patterns)) {
                throw new Error('event sheet patterns must be an array');
            }
            var count = { events: 0 };
            this.validateAnalysisStateEvents(sheet.events || [], 0, count, limits);
            return;
        }
        this.validateAnalysisStateGraphPayload(sheet.arg, 'arg', limits);
        this.validateAnalysisStateGraphPayload(sheet.concept, 'concept', limits);
    }

    validateAnalysisStateGraphPayload(graph, name, limits) {
        if (!graph) return;
        if (graph.elements != null && !Array.isArray(graph.elements)) {
            throw new Error(name + ' graph elements must be an array');
        }
        if (graph.elements && graph.elements.length > limits.maxGraphElements) {
            throw new Error(name + ' graph has too many elements');
        }
    }

    validateAnalysisStateEvents(events, depth, count, limits) {
        if (!Array.isArray(events)) {
            throw new Error('event children must be an array');
        }
        if (depth > limits.maxEventDepth) {
            throw new Error('event tree too deep');
        }
        for (var i = 0; i < events.length; i++) {
            count.events++;
            if (count.events > limits.maxEvents) {
                throw new Error('too many events in analysis state');
            }
            var ev = events[i] || {};
            if (ev.address != null && !/^\\d+(\\/\\d+)*$/.test(String(ev.address))) {
                throw new Error('invalid event address: ' + ev.address);
            }
            if (ev.subs != null) {
                this.validateAnalysisStateEvents(ev.subs, depth + 1, count, limits);
            }
        }
    }

    removeAnalysisStateExtraSheets() {
        var ids = Object.keys(this.sheets || {});
        for (var i = 0; i < ids.length; i++) {
            if (ids[i] !== 'sheet-1') {
                this.removeSheet(ids[i]);
            }
        }
    }

    async closeCurrentFile() {
        if (this._editorDirty()) {
            var choice = await this.showDirtyCloseDialog();
            if (choice === 'cancel') {
                this.controls.setStatus('Close cancelled');
                return;
            }
            if (choice === 'save') {
                var saved = await this.save();
                if (!saved) {
                    return;
                }
            }
        }
        await this.newModel({ skipSaveCurrent: true });
    }

    /**
     * Start a new model. Saves any existing state first, then clears everything.
     */
    async newModel(options) {
        options = options || {};
        // Save current state before clearing
        if (!options.skipSaveCurrent && this._persistedFileContent) {
            IvyPersist.save(this);
        }
        this._rememberLastOpenFile();

        // Create a fresh server session
        try {
            await this.api.createSession();
            this.updateSessionDisplay(IvyPersist.getSessionIdFromURL() || this.api.sessionId);
            IvyPersist.setSessionIdInURL(this.api.sessionId);

            // Reconnect SSE
            if (this.api.sessionId) {
                this.api.connectEvents(this.handleEvent.bind(this));
            }
        } catch (e) {
            this.controls.setStatus('Failed to create session: ' + e.message, 'error');
            return;
        }

        // Clear graphs
        this.argGraph.cy.elements().remove();
        this.conceptGraph.cy.elements().remove();

        // Clear state
        this._persistedFileName = '';
        this._persistedFilePath = '';
        this._persistedFileContent = '';
        this._persistedConceptRelations = null;
        this._fileHandle = null;
        this._savedFileContent = null;
        this.selectedArgNode = null;

        // Clear UI
        this.setEditorContent('');
        IvyPersist.setFileName('');
        this._updateReopenLastFileButton();
        if (window.__ivyVueBridge && typeof window.__ivyVueBridge.clearStateRelations === 'function') {
            window.__ivyVueBridge.clearStateRelations();
        } else {
            var tbody = document.getElementById('state-checkbox-body');
            if (tbody) tbody.innerHTML = '';
        }

        this.controls.setStatus('New model — load an .ivy file to begin', 'success');
    }

    /**
     * Populate the recent files list in the File dropdown.
     * Deduplicates by fileName+contentLength, keeping the most recent.
     * Shows truncated path context when file names collide.
     */
    populateRecentFiles() {
        var vueRecentFiles = window.__ivyVueBridge && typeof window.__ivyVueBridge.updateRecentFiles === 'function';
        var container = vueRecentFiles ? null : document.getElementById('file-recent-list');
        if (!vueRecentFiles && !container) return;
        if (!vueRecentFiles) {
            container.innerHTML = '';
        }

        var sessions = IvyPersist.listSessions();
        if (sessions.length === 0) {
            if (vueRecentFiles) {
                window.__ivyVueBridge.updateRecentFiles([], null);
                return;
            }
            var empty = document.createElement('a');
            empty.href = '#';
            empty.textContent = '(no recent files)';
            empty.style.color = '#666';
            empty.style.pointerEvents = 'none';
            container.appendChild(empty);
            return;
        }

        // Deduplicate: keep only the most recent entry per file path.
        // sessions is newest-first, so the first occurrence of each path is the most recent.
        var seen = {};
        var unique = [];
        for (var i = 0; i < sessions.length; i++) {
            var sess = sessions[i];
            if (!sess.fileName || sess.fileName === '(unnamed)') continue;
            var dedupKey = sess.filePath || sess.fileName;
            if (seen[dedupKey]) continue;
            seen[dedupKey] = true;
            unique.push(sess);
        }
        unique.sort(function (a, b) {
            return (a.fileName || '').localeCompare(b.fileName || '');
        });

        // Check for duplicate basenames to decide if path context is needed.
        var baseNameCount = {};
        for (var k = 0; k < unique.length; k++) {
            var bn = unique[k].fileName;
            baseNameCount[bn] = (baseNameCount[bn] || 0) + 1;
        }

        // Show up to 10 recent files
        var self = this;
        if (vueRecentFiles) {
            var items = [];
            for (var m = 0; m < unique.length && m < 10; m++) {
                var session = unique[m];
                var label = session.fileName;
                if (baseNameCount[session.fileName] > 1 && session.filePath) {
                    label = session.fileName + '  ' + IvyPersist.truncatePath(session.filePath, 15);
                }
                var title = '';
                if (session.timestamp) {
                    title = (session.filePath || session.fileName) + '\\nLast used: ' + new Date(session.timestamp).toLocaleString();
                }
                items.push({ id: session.id, label: label, title: title });
            }
            window.__ivyVueBridge.updateRecentFiles(items, function (id) {
                self.loadRecentSession(id);
            });
            return;
        }
        for (var j = 0; j < unique.length && j < 10; j++) {
            (function (s) {
                var displayName = s.fileName;
                // If basename appears more than once, show path context
                if (baseNameCount[s.fileName] > 1 && s.filePath) {
                    displayName = s.fileName + '  ' + IvyPersist.truncatePath(s.filePath, 15);
                }
                var link = document.createElement('a');
                link.href = '#';
                link.textContent = displayName;
                if (s.timestamp) {
                    var date = new Date(s.timestamp);
                    link.title = (s.filePath || s.fileName) + '\\nLast used: ' + date.toLocaleString();
                }
                link.addEventListener('click', function (e) {
                    e.preventDefault();
                    self.flashAndClose(this, function () { self.loadRecentSession(s.id); });
                });
                container.appendChild(link);
            })(unique[j]);
        }
    }

    /**
     * Load a recent session by its saved session ID.
     */
    async loadRecentSession(savedSessionId) {
        // Save current state first
        if (this._persistedFileContent) {
            IvyPersist.save(this);
        }

        var state = IvyPersist.loadSession(savedSessionId);
        if (!state || !state.fileContent) {
            this.controls.setStatus('Could not load session: no saved data', 'error');
            return;
        }

        // Create a fresh server session for this restore
        try {
            await this.api.createSession();
        } catch (e) {
            this.controls.setStatus('Failed to create session: ' + e.message, 'error');
            return;
        }

        // Reconnect SSE
        if (this.api.sessionId) {
            this.api.connectEvents(this.handleEvent.bind(this));
        }

        var restored = await IvyPersist.restore(this, state);
        if (restored) {
            // Keep the URL set by restore() (state.sessionId) — do NOT clobber it with
            // api.sessionId, which resets to s1/s2/... on every server restart and would
            // overwrite unrelated historical sessions stored under those same IDs.
            IvyPersist.setFileName(
                this._persistedFileName || state.fileName,
                this._persistedFilePath || state.filePath || state.fileName
            );
            this.updateSessionDisplay(IvyPersist.getSessionIdFromURL() || this.api.sessionId);
            this.controls.setStatus('Loaded: ' + state.fileName, 'success');
        } else {
            this.controls.setStatus('Restore failed', 'error');
        }
    }

    /**
     * Run a verification check in the currently selected mode.
     */
    async runCheck() {
        var mode = this.getMode();
        this.controls.showLoading('Running ' + mode + ' check...');
        this.controls.setStatus('Recompiling editor content...');
        try {
            // Always recompile from the current editor content so the
            // server checks exactly what the user sees, not a stale cache.
            var editorContent = this.cmEditor ? this.cmEditor.getValue() : this._persistedFileContent;
            if (editorContent) {
                await this.api.reloadContent(editorContent, this._persistedFileName || 'model.ivy');
            }
            this.controls.setStatus('Running ' + mode + ' check...');
            var result = await this.api.runCheck(mode);

            // After check: refresh ARG to show counterexample states.
            var argData = await this.api.getARG();
            if (argData && argData.elements) {
                this.argGraph.update(argData.elements, argData.positions);
            }

            // After check: refresh concept graph to pick up new abstract_value.
            // Matches Python: view_state() → set_parent_state() → recompute().
            var conceptData = await this.api.getConceptGraph();
            if (conceptData && conceptData.elements) {
                this._lastConceptData = conceptData;
                this.conceptGraph.update(conceptData.elements, conceptData.positions);
            }

            // Auto-check "+" for used relations.
            // Matches Python ivy_ui_cti.py show_used_relations():
            // checks "+" for any relation whose formula mentions constants from the CTI.
            if (result && result.used_relations) {
                await this._autoCheckUsedRelations(result.used_relations);
            }
            this.showCheckResult(result);
        } catch (e) {
            this.controls.setStatus('Check failed: ' + e.message, 'error');
            console.error('Check error:', e);
        } finally {
            this.controls.hideLoading();
        }
    }

    /**
     * Auto-check the "+" checkbox for used relations after finding a CTI.
     * Matches Python ivy_ui_cti.py show_used_relations → show_relation(rel, '+').
     * @param {Array<string>} relationNames - names of relations to auto-check
     */
    async _autoCheckUsedRelations(relationNames) {
        if (!relationNames || relationNames.length === 0) return;
        var usedSet = {};
        for (var i = 0; i < relationNames.length; i++) {
            usedSet[relationNames[i]] = true;
        }
        // Find checkbox rows and auto-check "+" for matching relations
        var tbody = document.getElementById('state-checkbox-body');
        if (!tbody) return;
        var rows = tbody.querySelectorAll('tr');
        for (var r = 0; r < rows.length; r++) {
            var nameCell = rows[r].querySelector('.name-col a');
            if (!nameCell) continue;
            var name = nameCell.textContent.trim();
            var baseName = name.split('(')[0];
            if (usedSet[name] || usedSet[baseName]) {
                // Auto-check the "+" checkbox (first checkbox, index 0)
                var inputs = rows[r].querySelectorAll('input[type="checkbox"]');
                if (inputs.length > 0 && !inputs[0].checked) {
                    inputs[0].checked = true;
                    // Trigger the toggle handler
                    await this.onEdgeToggle(name, 'all_to_all', true);
                }
            }
        }
    }

    /**
     * Display the result of a verification check.
     */
    showCheckResult(result) {
        if (!result) return;

        // The backend returns {status:"ok", result:"pass"/"fail"/"error", ...}.
        // Use result.result (the verification outcome), not result.status (HTTP status).
        var verdict = result.result || result.status;
        var z3note = result.z3_contacted ? ' [Z3: yes]' : ' [Z3: no]';
        var mode = result.mode ? ' (' + result.mode + ')' : '';

        if (verdict === 'pass') {
            this.controls.setStatus('Check PASSED' + mode + z3note, 'success');
            this.controls.showInfo('Verification Result', 'PASSED' + z3note + ': ' + (result.message || 'All properties hold.'));
        } else if (verdict === 'fail') {
            var failDetails = result.message || 'Counterexample found.';
            if (result.failed_conjecture) {
                failDetails += '\\n\\n' + result.failed_conjecture;
            }
            if (result.counterexample_details) {
                failDetails += '\\n\\n' + result.counterexample_details;
            } else if (result.counterexample_trace) {
                failDetails += '\\n\\nCounterexample trace:\\n' + result.counterexample_trace;
            }
            this.controls.setStatus('Check FAILED' + mode + z3note + ' - counterexample found', 'error');
            this.controls.showInfo('Verification Result', 'FAILED' + z3note + ': ' + failDetails);
            this.addCheckResultViewActions(result);
            if (result.arg) {
                this.argGraph.update(result.arg.elements, result.arg.positions);
            }
        } else if (verdict === 'error') {
            this.controls.setStatus('Check ERROR' + mode + z3note, 'error');
            this.controls.showInfo('Verification Error', result.message || 'Unknown error');
        } else {
            this.controls.setStatus('Check result: ' + verdict + z3note);
            this.controls.showInfo('Verification Result', result.message || JSON.stringify(result));
        }
    }

    addCheckResultViewActions(result) {
        if (!result || !result.trace_arg) return;
        var self = this;
        var openTrace = function () {
            self.openARGSheet('Error trace', result.trace_arg, result.trace_sheet_id);
        };

        var appendLegacyButton = function () {
            if (document.querySelector('[data-check-view-trace]')) return;
            var info = document.getElementById('info-content');
            if (!info) return;
            var button = document.createElement('button');
            button.type = 'button';
            button.className = 'btn small';
            button.setAttribute('data-check-view-trace', 'true');
            button.textContent = 'View error trace';
            button.addEventListener('click', openTrace);
            info.appendChild(document.createElement('br'));
            info.appendChild(button);
        };

        if (window.__ivyVueBridge && typeof window.__ivyVueBridge.setCheckTraceAction === 'function') {
            window.__ivyVueBridge.setCheckTraceAction(openTrace);
            return;
        }
        appendLegacyButton();
    }

    /**
     * Undo the last concept domain change.
     */
    async doUndo() {
        this.controls.setStatus('Undoing...');
        try {
            await this.api.undo();
            await this.refreshConceptGraph();
            this.controls.setStatus('Undo complete', 'success');
        } catch (e) {
            // "nothing to undo" is not an error — just a no-op.
            if (e.message && e.message.indexOf('nothing to undo') >= 0) {
                this.controls.setStatus('Nothing to undo');
            } else {
                this.controls.setStatus('Undo failed: ' + e.message, 'error');
            }
        }
    }

    /**
     * Reset the concept domain to initial state.
     */
    async resetDomain() {
        this.controls.setStatus('Resetting domain...');
        try {
            await this.api.resetDomain();
            await this.refreshConceptGraph();
            this.controls.setStatus('Domain reset', 'success');
        } catch (e) {
            this.controls.setStatus('Reset failed: ' + e.message, 'error');
        }
    }

    /**
     * Switch to diagram concept domain.
     */
    async diagramDomain() {
        this.controls.setStatus('Switching to diagram domain...');
        try {
            var result = await this.api.diagramDomain();
            if (result && result.concept) {
                this.updateConceptGraph(result.concept);
            } else {
                await this.refreshConceptGraph();
            }
            this.controls.setStatus('Diagram domain active', 'success');
        } catch (e) {
            this.controls.setStatus('Diagram domain failed: ' + e.message, 'error');
        }
    }

    /**
     * Refresh the concept graph from the server.
     */
    async refreshConceptGraph() {
        try {
            var result = await this.api.getConceptGraph(this.selectedArgNode);
            if (result && result.elements) {
                this.conceptGraph.update(result.elements, result.positions);
            }
            if (result) {
                this.populateStateCheckboxes(result);
            }
        } catch (e) {
            console.error('Concept graph refresh error:', e);
        }
    }

    // ================================================================
    // Server-Sent Events Handler
    // ================================================================

    /**
     * Handle an SSE event from the server.
     * Event types: arg_updated, concept_updated, status, check_result, error,
     *              toggles, proof_updated
     */
    handleEvent(event) {
        if (!event || !event.type) return;

        switch (event.type) {
            case 'arg_updated':
                if (event.data && event.data.elements) {
                    this.argGraph.update(event.data.elements, event.data.positions);
                }
                break;

            case 'concept_updated':
                if (event.data && event.data.elements) {
                    this.conceptGraph.update(event.data.elements, event.data.positions);
                }
                if (event.data && event.data.edge_names) {
                    this.controls.buildEdgeToggles(event.data.edge_names, this.onEdgeToggleChange.bind(this));
                }
                if (event.data && event.data.label_names) {
                    this.controls.buildLabelToggles(event.data.label_names, this.onLabelToggleChange.bind(this));
                }
                break;

            case 'status':
                if (event.data && event.data.message) {
                    this.controls.setStatus(event.data.message, event.data.level);
                }
                break;

            case 'check_result':
                this.controls.hideLoading();
                this.showCheckResult(event.data);
                break;

            case 'error':
                this.controls.hideLoading();
                this.controls.setStatus('Error: ' + (event.data.message || 'Unknown error'), 'error');
                if (event.data && event.data.message) {
                    this.controls.showInfo('Error', event.data.message);
                }
                break;

            case 'toggles':
                // Server is providing toggle configuration
                if (event.data && event.data.edge_names) {
                    this.controls.buildEdgeToggles(event.data.edge_names, this.onEdgeToggleChange.bind(this));
                }
                if (event.data && event.data.label_names) {
                    this.controls.buildLabelToggles(event.data.label_names, this.onLabelToggleChange.bind(this));
                }
                break;

            case 'proof_updated':
                // Future: update proof graph view
                break;

            case 'file_loaded':
                // File was loaded on the server; refresh graphs and state pane
                this.refreshAfterLoad(event.data);
                break;

            case 'check_started':
                this.controls.setStatus('Verification running...', 'info');
                break;

            case 'check_completed':
                var resultMsg = 'Check complete';
                if (event.data && event.data.result) {
                    resultMsg += ': ' + event.data.result;
                }
                this.controls.setStatus(resultMsg, 'success');
                break;

            case 'action_started':
                if (event.data && event.data.action) {
                    this.controls.setStatus('Running: ' + event.data.action + '...', 'info');
                }
                break;

            case 'action_completed':
                if (event.data && event.data.action) {
                    if (event.data.status && event.data.status !== 'ok') {
                        this.controls.setStatus('Action failed: ' + event.data.status, 'error');
                    } else {
                        this.controls.setStatus('Done: ' + event.data.action, 'success');
                    }
                }
                break;

            default:
                // Silently ignore unknown event types
                break;
        }
    }

    // ================================================================
    // Dropdown Menu Infrastructure
    // ================================================================

    /**
     * Set up panel header dropdown menus (click to toggle).
     */
    setupDropdownMenus() {
        var self = this;
        var dropdowns = document.querySelectorAll('.dropdown > .panel-menu');
        for (var i = 0; i < dropdowns.length; i++) {
            (function (trigger) {
                trigger.addEventListener('click', function (e) {
                    e.stopPropagation();
                    var parent = trigger.parentElement;
                    var wasOpen = parent.classList.contains('open');
                    // Close all dropdowns first
                    var all = document.querySelectorAll('.dropdown.open');
                    for (var j = 0; j < all.length; j++) {
                        all[j].classList.remove('open');
                    }
                    if (!wasOpen) {
                        parent.classList.add('open');
                        // Populate recent files when File menu opens
                        var dropdownId = trigger.getAttribute('data-dropdown');
                        if (dropdownId === 'file-menu') {
                            self.populateRecentFiles();
                        }
                    }
                });
            })(dropdowns[i]);
        }
    }

    /**
     * Close all open dropdown menus.
     */
    closeAllDropdowns() {
        if (window.__ivyVueBridge && typeof window.__ivyVueBridge.closeDropdownMenus === 'function') {
            window.__ivyVueBridge.closeDropdownMenus();
            return;
        }
        var all = document.querySelectorAll('.dropdown.open');
        for (var j = 0; j < all.length; j++) {
            all[j].classList.remove('open');
        }
    }

    /**
     * Flash a menu item (macOS Cocoa style invert) then close dropdowns and invoke callback.
     */
    flashAndClose(el, callback) {
        var self = this;
        var bridge = window.__ivyVueBridge;
        if (bridge && typeof bridge.flashMenuItem === 'function' && el && el.id) {
            bridge.flashMenuItem(el.id, 50);
            setTimeout(function () {
                self.closeAllDropdowns();
                if (callback) callback();
            }, 50);
            return;
        }
        el.classList.add('menu-flash');
        setTimeout(function () {
            el.classList.remove('menu-flash');
            self.closeAllDropdowns();
            if (callback) callback();
        }, 50);
    }

    /**
     * Bind a menu item by ID to a callback, with dropdown auto-close.
     */
    bindMenuAction(id, callback) {
        var self = this;
        var el = document.getElementById(id);
        if (!el) return;
        el.addEventListener('click', function (e) {
            e.preventDefault();
            e.stopPropagation();
            self.flashAndClose(this, callback);
        });
    }

    async loadMenuDescriptors() {
        try {
            var menus = await this.api.getMenus();
            this.renderMenuRegion('arg', menus.arg || []);
            this.renderMenuRegion('concept', menus.concept || []);
        } catch (e) {
            this.controls.setStatus('Menu load failed: ' + e.message, 'error');
        }
    }

    renderMenuRegion(region, menus) {
        if (window.__ivyVueBridge && typeof window.__ivyVueBridge.updateMenuRegion === 'function') {
            var self = this;
            window.__ivyVueBridge.updateMenuRegion(region, menus || [], function (item) {
                return self.dispatchMenuDescriptorAction(region, item);
            });
            return;
        }
        var panel = region === 'arg' ? document.getElementById('arg-panel') : document.getElementById('concept-panel');
        if (!panel) return;
        var header = panel.querySelector('.panel-header');
        if (!header) return;
        var old = header.querySelector('[data-dynamic-menu-region="' + region + '"]');
        if (old) old.remove();
        var menuRow = header.querySelector('.panel-header-actions') || header;

        var root = document.createElement('div');
        root.className = 'dynamic-menu-root';
        root.setAttribute('data-dynamic-menu-region', region);
        menuRow.appendChild(root);

        for (var i = 0; i < menus.length; i++) {
            this.renderMenuDescriptor(root, region, menus[i], i);
        }
    }

    renderMenuDescriptor(root, region, menu, index) {
        var self = this;
        var dropdown = document.createElement('div');
        dropdown.className = 'dropdown';

        var contentId = 'dynamic-' + region + '-' + index + '-' + (menu.label || 'menu').toLowerCase().replace(/[^a-z0-9]+/g, '-');
        var label = document.createElement('span');
        label.className = 'panel-menu';
        label.textContent = menu.label || 'Menu';
        label.setAttribute('data-dropdown', contentId);
        label.addEventListener('click', function (e) {
            e.preventDefault();
            e.stopPropagation();
            var wasOpen = dropdown.classList.contains('open');
            self.closeAllDropdowns();
            if (!wasOpen) dropdown.classList.add('open');
        });
        dropdown.appendChild(label);

        var content = document.createElement('div');
        content.id = contentId;
        content.className = 'dropdown-content';
        dropdown.appendChild(content);

        var items = menu.items || [];
        for (var i = 0; i < items.length; i++) {
            var item = items[i];
            if (item.type === 'separator') {
                var sep = document.createElement('div');
                sep.className = 'dropdown-sep';
                content.appendChild(sep);
                continue;
            }
            var link = document.createElement('a');
            link.href = '#';
            link.textContent = item.label || item.action || '';
            link.setAttribute('data-menu-action', item.action || '');
            link.setAttribute('data-menu-dispatch', item.dispatch || '');
            if (item.enabled === false) {
                link.classList.add('disabled');
                link.setAttribute('aria-disabled', 'true');
            }
            link.addEventListener('click', function (e) {
                e.preventDefault();
                e.stopPropagation();
                var actionItem = this.__ivyMenuItem;
                self.closeAllDropdowns();
                self.dispatchMenuDescriptorAction(region, actionItem);
            });
            link.__ivyMenuItem = item;
            content.appendChild(link);
        }

        root.appendChild(dropdown);
    }

    dispatchMenuDescriptorAction(region, item) {
        if (!item || item.enabled === false) return Promise.resolve({ ok: false, error: 'disabled action' });
        if (item.dispatch === 'action') {
            return this.runAction(item.action, {}, {
                runningMessage: 'Running: ' + item.action + '...',
                successMessage: 'Done: ' + item.action,
            });
        }
        return this.runAction(item.action, {});
    }

    /**
     * Shared browser action runner, matching Python's run_context shape:
     * show progress, surface backend errors, and return a structured outcome.
     */
    async runAction(actionName, args, options) {
        var opts = options || {};
        var runningMessage = opts.runningMessage || ('Running: ' + actionName + '...');
        var successMessage = opts.successMessage || ('Done: ' + actionName);
        var failurePrefix = opts.failurePrefix || 'Action failed';
        if (opts.showLoading !== false) {
            this.controls.showLoading(runningMessage);
        }
        this.controls.setStatus(runningMessage, 'info');
        try {
            var result = await this.api.executeAction(actionName, args || {});
            if (opts.successStatus !== false) {
                this.controls.setStatus(successMessage, 'success');
            }
            return { ok: true, result: result };
        } catch (e) {
            var message = failurePrefix + ': ' + e.message;
            this.controls.setStatus(message, 'error');
            if (opts.showDialog) {
                this.showTextDialog('ivyweb', failurePrefix, e.message);
            }
            return { ok: false, error: e.message };
        } finally {
            if (opts.showLoading !== false) {
                this.controls.hideLoading();
            }
        }
    }

    // ================================================================
    // Verification Operations (Invariant menu)
    // Matches Python ivy_ui_cti.py
    // ================================================================

    /**
     * Check inductiveness of current conjectures.
     * Matches Python ivy_ui_cti.py check_inductiveness().
     */
    async checkInduction() {
        this.controls.setStatus('Checking induction...');
        this.controls.showLoading('Checking inductiveness...');
        try {
            var result = await this.api.runCheck('induction');
            if (result.result === 'fail' && result.failed_conjecture) {
                // Show dialog matching Python ivy_ui_cti.py:
                // "The following conjecture is not relatively inductive:"
                this.showTextDialog(
                    'ivyweb',
                    result.message || 'The following conjecture is not relatively inductive:',
                    result.failed_conjecture
                );
                this.controls.setStatus('Induction check: not inductive');
            } else if (result.result === 'pass') {
                // Success — show the invariant in a dialog
                this.showTextDialog(
                    'ivyweb',
                    'Inductive invariant found:',
                    result.message.replace('Inductive invariant found:\\n', '')
                );
                this.controls.setStatus('Induction check: PASSED', 'success');
            } else {
                this.controls.setStatus('Induction check: ' + (result.message || result.result));
            }
        } catch (e) {
            this.controls.setStatus('Induction check failed: ' + e.message, 'error');
        } finally {
            this.controls.hideLoading();
        }
    }

    _createDialog(title, message) {
        var overlay = document.createElement('div');
        overlay.className = 'dialog-overlay';
        overlay.setAttribute('data-ivy-dialog', 'true');
        overlay.style.display = 'flex';

        var box = document.createElement('div');
        box.className = 'dialog-box';
        overlay.appendChild(box);

        var titleEl = document.createElement('div');
        titleEl.className = 'dialog-title';
        titleEl.textContent = title || 'ivyweb';
        box.appendChild(titleEl);

        var messageEl = document.createElement('div');
        messageEl.className = 'dialog-message';
        messageEl.textContent = message || '';
        box.appendChild(messageEl);

        var body = document.createElement('div');
        body.className = 'dialog-body';
        box.appendChild(body);

        var error = document.createElement('div');
        error.className = 'dialog-message dialog-error';
        error.setAttribute('data-ivy-dialog-error', 'true');
        error.style.display = 'none';
        box.appendChild(error);

        var buttons = document.createElement('div');
        buttons.className = 'dialog-buttons';
        box.appendChild(buttons);

        document.body.appendChild(overlay);
        return { overlay: overlay, body: body, buttons: buttons, error: error };
    }

    _setDialogError(dialog, message) {
        dialog.error.textContent = message || '';
        dialog.error.style.display = message ? 'block' : 'none';
    }

    _addDialogButton(dialog, label, callback, extraClass) {
        var btn = document.createElement('button');
        btn.type = 'button';
        btn.className = 'dialog-btn' + (extraClass ? ' ' + extraClass : '');
        btn.textContent = label || 'OK';
        btn.setAttribute('data-ivy-dialog-button', 'true');
        btn.addEventListener('click', callback);
        dialog.buttons.appendChild(btn);
        return btn;
    }

    _finishDialog(dialog, cleanup, resolve, value) {
        if (cleanup) cleanup();
        dialog.overlay.remove();
        resolve(value);
    }

    _installDialogEscape(dialog, finish) {
        var handler = function (e) {
            if (e.key === 'Escape') {
                finish();
            }
        };
        document.addEventListener('keydown', handler);
        return function () {
            document.removeEventListener('keydown', handler);
        };
    }

    okDialog(title, message) {
        if (window.__ivyVueBridge && typeof window.__ivyVueBridge.showDialog === 'function') {
            return window.__ivyVueBridge.showDialog({ type: 'ok', title: title, message: message });
        }
        var self = this;
        return new Promise(function (resolve) {
            var dialog = self._createDialog(title, message);
            var cleanup = self._installDialogEscape(dialog, function () {
                self._finishDialog(dialog, cleanup, resolve, true);
            });
            self._addDialogButton(dialog, 'OK', function () {
                self._finishDialog(dialog, cleanup, resolve, true);
            });
        });
    }

    okCancelDialog(title, message) {
        if (window.__ivyVueBridge && typeof window.__ivyVueBridge.showDialog === 'function') {
            return window.__ivyVueBridge.showDialog({ type: 'okCancel', title: title, message: message });
        }
        var self = this;
        return new Promise(function (resolve) {
            var dialog = self._createDialog(title, message);
            var cleanup = self._installDialogEscape(dialog, function () {
                self._finishDialog(dialog, cleanup, resolve, false);
            });
            self._addDialogButton(dialog, 'Cancel', function () {
                self._finishDialog(dialog, cleanup, resolve, false);
            });
            self._addDialogButton(dialog, 'OK', function () {
                self._finishDialog(dialog, cleanup, resolve, true);
            });
        });
    }

    textDialog(title, message, text, options) {
        if (window.__ivyVueBridge && typeof window.__ivyVueBridge.showDialog === 'function') {
            return window.__ivyVueBridge.showDialog({
                type: 'text',
                title: title,
                message: message,
                text: text,
                options: options || {},
            });
        }
        var self = this;
        var opts = options || {};
        return new Promise(function (resolve) {
            var dialog = self._createDialog(title, message);
            var textarea = document.createElement('textarea');
            textarea.className = 'dialog-text';
            textarea.rows = opts.rows || 4;
            textarea.cols = opts.cols || 80;
            textarea.value = text || '';
            textarea.readOnly = !!opts.readOnly;
            textarea.setAttribute('data-ivy-dialog-text', 'true');
            dialog.body.appendChild(textarea);

            var cleanup = self._installDialogEscape(dialog, function () {
                self._finishDialog(dialog, cleanup, resolve, null);
            });
            if (opts.cancel) {
                self._addDialogButton(dialog, 'Cancel', function () {
                    self._finishDialog(dialog, cleanup, resolve, null);
                });
            }
            self._addDialogButton(dialog, opts.okLabel || 'OK', function () {
                self._finishDialog(dialog, cleanup, resolve, textarea.value);
            });
            textarea.focus();
            textarea.select();
        });
    }

    /**
     * Backward-compatible display-only text dialog.
     */
    showTextDialog(title, message, text) {
        this.textDialog(title, message, text, { readOnly: false, okLabel: 'OK' });
    }

    entryDialog(title, message, initialValue, options) {
        if (window.__ivyVueBridge && typeof window.__ivyVueBridge.showDialog === 'function') {
            return window.__ivyVueBridge.showDialog({
                type: 'entry',
                title: title,
                message: message,
                initialValue: initialValue,
                options: options || {},
            });
        }
        var self = this;
        var opts = options || {};
        return new Promise(function (resolve) {
            var dialog = self._createDialog(title, message);
            var input = document.createElement('input');
            input.type = 'text';
            input.className = 'dialog-input';
            input.value = initialValue || '';
            input.setAttribute('data-ivy-dialog-entry', 'true');
            dialog.body.appendChild(input);

            var cleanup = self._installDialogEscape(dialog, function () {
                self._finishDialog(dialog, cleanup, resolve, null);
            });
            if (opts.cancel !== false) {
                self._addDialogButton(dialog, 'Cancel', function () {
                    self._finishDialog(dialog, cleanup, resolve, null);
                });
            }
            self._addDialogButton(dialog, opts.okLabel || 'OK', function () {
                self._finishDialog(dialog, cleanup, resolve, input.value);
            });
            input.focus();
            input.select();
        });
    }

    integerDialog(title, message, initialValue, options) {
        if (window.__ivyVueBridge && typeof window.__ivyVueBridge.showDialog === 'function') {
            return window.__ivyVueBridge.showDialog({
                type: 'integer',
                title: title,
                message: message,
                initialValue: initialValue,
                options: options || {},
            });
        }
        var self = this;
        var opts = options || {};
        return new Promise(function (resolve) {
            var dialog = self._createDialog(title, message);
            var input = document.createElement('input');
            input.type = 'number';
            input.className = 'dialog-input';
            input.value = String(initialValue == null ? '' : initialValue);
            if (opts.min != null) input.min = String(opts.min);
            if (opts.max != null) input.max = String(opts.max);
            input.setAttribute('data-ivy-dialog-int', 'true');
            dialog.body.appendChild(input);

            var cleanup = self._installDialogEscape(dialog, function () {
                self._finishDialog(dialog, cleanup, resolve, null);
            });
            if (opts.cancel !== false) {
                self._addDialogButton(dialog, 'Cancel', function () {
                    self._finishDialog(dialog, cleanup, resolve, null);
                });
            }
            self._addDialogButton(dialog, opts.okLabel || 'OK', function () {
                var raw = input.value.trim();
                var value = Number(raw);
                if (raw === '' || !Number.isInteger(value)) {
                    self._setDialogError(dialog, 'Enter an integer.');
                    return;
                }
                if (opts.min != null && value < opts.min) {
                    self._setDialogError(dialog, 'Enter a value at least ' + opts.min + '.');
                    return;
                }
                if (opts.max != null && value > opts.max) {
                    self._setDialogError(dialog, 'Enter a value at most ' + opts.max + '.');
                    return;
                }
                self._finishDialog(dialog, cleanup, resolve, value);
            });
            input.focus();
            input.select();
        });
    }

    listboxDialog(title, message, items, options) {
        if (window.__ivyVueBridge && typeof window.__ivyVueBridge.showDialog === 'function') {
            return window.__ivyVueBridge.showDialog({
                type: 'listbox',
                title: title,
                message: message,
                items: items || [],
                options: options || {},
            });
        }
        var self = this;
        var opts = options || {};
        var entries = (items || []).map(function (item) {
            if (typeof item === 'object' && item !== null) {
                return { label: item.label || String(item.value), value: item.value };
            }
            return { label: String(item), value: item };
        });
        return new Promise(function (resolve) {
            var dialog = self._createDialog(title, message);
            var select = document.createElement('select');
            select.className = 'dialog-input dialog-listbox';
            select.size = opts.size || Math.min(Math.max(entries.length, 2), 12);
            select.multiple = !!opts.multiple;
            select.setAttribute('data-ivy-dialog-list', 'true');
            entries.forEach(function (entry, index) {
                var opt = document.createElement('option');
                opt.value = String(entry.value);
                opt.setAttribute('data-ivy-dialog-index', String(index));
                opt.textContent = entry.label;
                select.appendChild(opt);
            });
            dialog.body.appendChild(select);

            var cleanup = self._installDialogEscape(dialog, function () {
                self._finishDialog(dialog, cleanup, resolve, opts.multiple ? [] : null);
            });
            if (opts.cancel !== false) {
                self._addDialogButton(dialog, 'Cancel', function () {
                    self._finishDialog(dialog, cleanup, resolve, opts.multiple ? [] : null);
                });
            }
            self._addDialogButton(dialog, opts.okLabel || 'OK', function () {
                if (opts.multiple) {
                    var selected = Array.from(select.selectedOptions).map(function (opt) {
                        return entries[Number(opt.getAttribute('data-ivy-dialog-index'))].value;
                    });
                    self._finishDialog(dialog, cleanup, resolve, selected);
                    return;
                }
                var selectedOption = select.selectedOptions[0];
                var idx = selectedOption ? Number(selectedOption.getAttribute('data-ivy-dialog-index')) : -1;
                self._finishDialog(dialog, cleanup, resolve, idx >= 0 ? entries[idx].value : null);
            });
            select.focus();
        });
    }

    buttonListDialog(title, message, buttons) {
        if (window.__ivyVueBridge && typeof window.__ivyVueBridge.showDialog === 'function') {
            return window.__ivyVueBridge.showDialog({
                type: 'buttons',
                title: title,
                message: message,
                buttons: buttons || [],
            });
        }
        var self = this;
        var entries = buttons || [];
        return new Promise(function (resolve) {
            var dialog = self._createDialog(title, message);
            var cleanup = self._installDialogEscape(dialog, function () {
                self._finishDialog(dialog, cleanup, resolve, null);
            });
            entries.forEach(function (entry) {
                var label = entry.label || String(entry.value);
                self._addDialogButton(dialog, label, function () {
                    self._finishDialog(dialog, cleanup, resolve, entry.value);
                }, entry.danger ? 'dialog-btn-danger' : '');
            });
        });
    }

    /**
     * Run bounded model checking.
     * Matches Python ivy_ui_cti.py bounded_check().
     */
    async boundedCheck() {
        try {
            var bound = await this.integerDialog('Bounded check', 'Enter bound:', this.currentBound, {
                min: 1,
                okLabel: 'OK',
            });
            if (bound === null) {
                this.controls.setStatus('Bounded check cancelled');
                return;
            }
            this.currentBound = bound;
            this.controls.setStatus('Running bounded check...');
            var result = await this.api.runCheck('bounded', { bound: bound });
            this.controls.setStatus('Bounded check: ' + (result.result || 'done'), 'success');
        } catch (e) {
            this.controls.setStatus('Bounded check failed: ' + e.message, 'error');
        }
    }

    /**
     * Weaken the current invariant.
     * Matches Python ivy_ui_cti.py weaken().
     */
    async weakenInvariant() {
        try {
            this.controls.setStatus('Choosing conjectures...');
            var choicesResult = await this.api.executeAction('get_conjectures', {});
            var conjectures = (choicesResult && choicesResult.conjectures) || [];
            var choices = conjectures.map(function (conj, index) {
                var label = conj.label ? '[' + conj.label + '] ' + conj.formula : conj.formula;
                return { label: label || String(index), value: index };
            });
            var indices = await this.listboxDialog('Weaken', 'Choose conjectures to remove:', choices, {
                multiple: true,
                okLabel: 'Weaken',
            });
            if (!indices || indices.length === 0) {
                this.controls.setStatus('Weaken cancelled');
                return;
            }
            this.controls.setStatus('Weakening invariant...');
            var result = await this.api.executeAction('weaken', { indices: indices });
            await this.refreshConceptGraph();
            this.controls.setStatus('Invariant weakened', 'success');
        } catch (e) {
            this.controls.setStatus('Weaken failed: ' + e.message, 'error');
        }
    }

    /**
     * Save the current abstraction to a file.
     * Matches Python ivy_ui.py save_abstraction().
     */
    async saveAbstraction() {
        this.controls.setStatus('Saving abstraction...');
        try {
            // Build abstraction content from concept session state
            var result = await this.api.executeAction('save_abstraction', {});
            var content = (result && result.content) || '';
            if (!content) {
                // Fallback: serialize the concept graph elements as text
                if (this.conceptGraph && this.conceptGraph.cy) {
                    var elements = this.conceptGraph.cy.json().elements;
                    content = '# Ivy concept abstraction\\n# Generated by Ivy web UI\\n\\n';
                    content += JSON.stringify(elements, null, 2) + '\\n';
                } else {
                    content = '# (no abstraction data available)\\n';
                }
            }

            var suggestedName = (this._persistedFileName || 'abstraction').replace(/\\.ivy$/, '') + '_abstraction.ivy';

            if (window.showSaveFilePicker) {
                var handle = await window.showSaveFilePicker({
                    suggestedName: suggestedName,
                    types: [{
                        description: 'Ivy files',
                        accept: { 'text/plain': ['.ivy'] },
                    }],
                });
                var writable = await handle.createWritable();
                await writable.write(content);
                await writable.close();
                this.controls.setStatus('Abstraction saved: ' + handle.name, 'success');
            } else {
                // Fallback: browser download
                var blob = new Blob([content], { type: 'text/plain' });
                var url = URL.createObjectURL(blob);
                var a = document.createElement('a');
                a.href = url;
                a.download = suggestedName;
                document.body.appendChild(a);
                a.click();
                document.body.removeChild(a);
                URL.revokeObjectURL(url);
                this.controls.setStatus('Abstraction downloaded: ' + suggestedName, 'success');
            }
        } catch (e) {
            if (e.name === 'AbortError') {
                this.controls.setStatus('Save abstraction cancelled');
            } else {
                this.controls.setStatus('Save abstraction failed: ' + e.message, 'error');
            }
        }
    }

    // ================================================================
    // Conjecture Menu Operations
    // Matches Python ivy_graph_ui.py GraphWidget menus
    // ================================================================

    /**
     * Redo the last undone operation.
     */
    async doRedo() {
        this.controls.setStatus('Redo...');
        try {
            await this.api.executeAction('redo', {});
            await this.refreshConceptGraph();
            this.controls.setStatus('Redo complete', 'success');
        } catch (e) {
            this.controls.setStatus('Redo failed: ' + e.message, 'error');
        }
    }

    /**
     * Perform one step of PDR strengthening.
     * Matches Python ivy_graph_ui.py pdr_step().
     */
    async pdrStep() {
        this.controls.setStatus('PDR step...');
        try {
            var result = await this.api.executeAction('pdr_step', {});
            await this.refreshConceptGraph();
            this.controls.setStatus('PDR step complete', 'success');
        } catch (e) {
            this.controls.setStatus('PDR step failed: ' + e.message, 'error');
        }
    }

    async showReachableStates() {
        this.controls.setStatus('Opening reachable states...');
        try {
            var result = await this.api.executeAction('show_reachable', {});
            this.openARGSheet('Reachable states', result.arg, result.sheet_id);
            this.controls.setStatus('Reachable states opened', 'success');
        } catch (e) {
            this.controls.setStatus('Show reachable states failed: ' + e.message, 'error');
        }
    }

    /**
     * Show concrete model.
     * Matches Python ivy_graph_ui.py concrete().
     */
    async concreteStep() {
        this.controls.setStatus('Computing concrete model...');
        try {
            var result = await this.api.executeAction('concrete', { sheet_id: this.activeSheetId || 'sheet-1' });
            if (result && result.concept) {
                this.updateConceptGraph(result.concept);
            } else {
                await this.refreshConceptGraph();
            }
            this.controls.setStatus('Concrete model computed', 'success');
        } catch (e) {
            this.controls.setStatus('Concrete failed: ' + e.message, 'error');
        }
    }

    /**
     * Gather facts from the current state.
     * Matches Python ivy_graph_ui.py gather().
     */
    async gatherFacts() {
        this.controls.setStatus('Gathering facts...');
        try {
            var result = await this.api.executeAction('gather', { sheet_id: this.activeSheetId || 'sheet-1' });
            if (result && result.concept) {
                this.updateConceptGraph(result.concept);
            } else {
                await this.refreshConceptGraph();
            }
            this.controls.setStatus('Facts gathered', 'success');
        } catch (e) {
            this.controls.setStatus('Gather failed: ' + e.message, 'error');
        }
    }

    async ctiConceptAction(actionName) {
        this.controls.setStatus('Running CTI action...');
        try {
            var result = await this.api.executeAction(actionName, { sheet_id: this.activeSheetId || 'sheet-1' });
            if (result && result.concept) {
                this.updateConceptGraph(result.concept);
            } else {
                await this.refreshConceptGraph();
            }
            this.controls.setStatus((result && result.message) || 'CTI action complete', 'success');
        } catch (e) {
            this.controls.setStatus('CTI action failed: ' + e.message, 'error');
        }
    }

    /**
     * Compute reverse image.
     * Matches Python ivy_graph_ui.py reverse().
     */
    async reverseStep() {
        this.controls.setStatus('Computing reverse...');
        try {
            var result = await this.api.executeAction('reverse', {});
            await this.refreshConceptGraph();
            this.controls.setStatus('Reverse complete', 'success');
        } catch (e) {
            this.controls.setStatus('Reverse failed: ' + e.message, 'error');
        }
    }

    /**
     * Compute reachable states along a path.
     * Matches Python ivy_graph_ui.py path_reach().
     */
    async pathReach() {
        this.controls.setStatus('Computing path reachability...');
        try {
            var result = await this.api.executeAction('path_reach', { sheet_id: this.activeSheetId || 'sheet-1' });
            await this.refreshConceptGraph();
            this.controls.setStatus('Path reach complete', 'success');
        } catch (e) {
            this.controls.setStatus('Path reach failed: ' + e.message, 'error');
        }
    }

    /**
     * Compute reachable states.
     * Matches Python ivy_graph_ui.py reach().
     */
    async reachStep() {
        this.controls.setStatus('Computing reachability...');
        try {
            var result = await this.api.executeAction('reach', { sheet_id: this.activeSheetId || 'sheet-1' });
            await this.refreshConceptGraph();
            this.controls.setStatus('Reach complete', 'success');
        } catch (e) {
            this.controls.setStatus('Reach failed: ' + e.message, 'error');
        }
    }

    /**
     * Generate a conjecture from the current state.
     * Matches Python ivy_graph_ui.py conjecture().
     */
    async makeConjecture() {
        this.controls.setStatus('Generating conjecture...');
        try {
            var result = await this.api.executeAction('conjecture', {});
            await this.refreshConceptGraph();
            this.controls.setStatus('Conjecture generated', 'success');
        } catch (e) {
            this.controls.setStatus('Conjecture failed: ' + e.message, 'error');
        }
    }

    /**
     * Backtrack to a previous checkpoint.
     * Matches Python ivy_graph_ui.py backtrack().
     */
    async backtrack() {
        this.controls.setStatus('Backtracking...');
        try {
            await this.api.executeAction('backtrack', { sheet_id: this.activeSheetId || 'sheet-1' });
            await this.refreshConceptGraph();
            this.controls.setStatus('Backtracked', 'success');
        } catch (e) {
            this.controls.setStatus('Backtrack failed: ' + e.message, 'error');
        }
    }

    /**
     * Recalculate the current concept graph.
     * Matches Python ivy_graph_ui.py recalculate().
     */
    async recalculateGraph() {
        this.controls.setStatus('Recalculating...');
        try {
            await this.api.executeAction('recalculate', {});
            await this.refreshConceptGraph();
            this.controls.setStatus('Recalculated', 'success');
        } catch (e) {
            this.controls.setStatus('Recalculate failed: ' + e.message, 'error');
        }
    }

    /**
     * Remember the current graph for later use.
     * Matches Python ivy_graph_ui.py remember().
     */
    async rememberGraph() {
        this.controls.setStatus('Remembering graph...');
        try {
            var name = await this.entryDialog('Remember graph', 'Enter a name for this goal:', '', { okLabel: 'Remember' });
            if (name === null) {
                this.controls.setStatus('Remember cancelled', 'warning');
                return;
            }
            await this.api.executeAction('remember', { name: name, sheet_id: this.activeSheetId || 'sheet-1' });
            this.controls.setStatus('Graph remembered', 'success');
        } catch (e) {
            this.controls.setStatus('Remember failed: ' + e.message, 'error');
        }
    }

    /**
     * Export the current conjecture.
     * Matches Python ivy_graph_ui.py export().
     */
    async exportConjecture() {
        this.controls.setStatus('Exporting conjecture...');
        try {
            var result = await this.api.executeAction('export', { sheet_id: this.activeSheetId || 'sheet-1' });
            var content = (result && result.content) || '';
            var filename = (result && result.filename) || 'concept_graph.dot';
            if (content) {
                if (window.showSaveFilePicker) {
                    var handle = await window.showSaveFilePicker({
                        suggestedName: filename,
                        types: [{
                            description: 'DOT files',
                            accept: { 'text/vnd.graphviz': ['.dot'] },
                        }],
                    });
                    var writable = await handle.createWritable();
                    await writable.write(content);
                    await writable.close();
                } else {
                    var blob = new Blob([content], { type: (result && result.mime_type) || 'text/vnd.graphviz' });
                    var url = URL.createObjectURL(blob);
                    var a = document.createElement('a');
                    a.href = url;
                    a.download = filename;
                    document.body.appendChild(a);
                    a.click();
                    document.body.removeChild(a);
                    URL.revokeObjectURL(url);
                }
            }
            this.controls.setStatus('Graph exported', 'success');
        } catch (e) {
            this.controls.setStatus('Export failed: ' + e.message, 'error');
        }
    }

    /**
     * Add a relation from a user-entered string.
     * Matches Python ivy_graph_ui.py add_concept_from_string().
     */
    async addRelationFromString() {
        var input = await this.entryDialog(
            'Add relation',
            'Add a relation [example: p(X,a,Y)]:',
            '',
            { okLabel: 'Add' }
        );
        if (!input) return;
        try {
            await this.api.executeAction('add_relation', { formula: input });
            await this.refreshConceptGraph();
            this.controls.setStatus('Relation added', 'success');
        } catch (e) {
            this.controls.setStatus('Add relation failed: ' + e.message, 'error');
        }
    }

    /**
     * Refresh graphs and state pane after a file_loaded SSE event.
     * The event.data contains {filename, sorts, relations, actions}.
     */
    async refreshAfterLoad(data) {
        try {
            // Refresh ARG
            var argData = await this.api.getARG();
            if (argData && argData.elements) {
                this.argGraph.update(argData.elements, argData.positions);
            }
            // Refresh concept graph and populate state checkbox pane
            var conceptData = await this.api.getConceptGraph();
            if (conceptData && conceptData.elements) {
                this.conceptGraph.update(conceptData.elements, conceptData.positions);
            }
            this.populateStateCheckboxes(conceptData);
            this.updateStateLabel(0);
        } catch (e) {
            console.error('refreshAfterLoad error:', e);
        }
    }

    /**
     * Show a non-modal toast notification. Auto-dismisses after 10s
     * or on click. Appends to document body so it floats over everything.
     */
    _showToast(message, level, options) {
        options = options || {};
        if (window.__ivyVueBridge && typeof window.__ivyVueBridge.showToast === 'function') {
            return window.__ivyVueBridge.showToast(message, level, options);
        }
        var toast = document.createElement('div');
        toast.className = 'ivy-toast ivy-toast-floating ivy-toast-' + (level || 'info');
        if (options.className) {
            toast.className += ' ' + options.className;
        }
        toast.textContent = message;
        toast.onclick = function () { toast.remove(); };
        document.body.appendChild(toast);
        if (!options.persistent) {
            setTimeout(function () { toast.remove(); }, 10000);
        }
        return toast;
    }
}

// ================================================================
// Initialize on DOM ready unless a framework shell owns boot timing.
// ================================================================
function startIvyApp() {
    if (window.ivyApp) {
        return window.ivyApp;
    }
    window.ivyApp = new IvyApp();
    window.ivyApp.init();
    return window.ivyApp;
}

window.IvyApp = IvyApp;
window.startIvyApp = startIvyApp;

if (!window.__IVY_VUE_OWNS_BOOT__) {
    document.addEventListener('DOMContentLoaded', startIvyApp);
}
`;function Nu({win:e=globalThis.window,source:t=Pu}={}){if(!e||typeof e.startIvyApp=="function")return e&&e.startIvyApp;e.__IVY_VUE_OWNS_BOOT__=!0;const n=e.eval||globalThis.eval;if(typeof n!="function")throw new Error("Unable to install IvyApp runtime: global eval is unavailable");return n.call(e,`${t}
//# sourceURL=ivyweb_app.compat.js`),e.startIvyApp}const Bu=[{selector:"node",style:{content:"data(label)","text-wrap":"wrap","font-size":"14px","text-outline-width":"0px","text-valign":"center",color:"#000",width:"data(width)",height:"data(height)","border-color":"#000",shape:"data(shape)","background-color":"#fff"}},{selector:"node.non_existing",style:{display:"none"}},{selector:"node.exactly_one",style:{"border-width":"4px","border-style":"solid","border-color":"#000"}},{selector:"node.at_least_one",style:{"border-width":"8px","border-style":"double","border-color":"#000"}},{selector:"node.at_most_one",style:{"border-width":"3px","border-style":"dotted","border-color":"#000"}},{selector:"node.node_unknown",style:{"border-width":"5px","border-style":"double","border-color":"#000"}},{selector:"edge",style:{width:"3px","line-color":"#888","target-arrow-color":"#888","source-arrow-color":"#888","target-arrow-shape":"triangle","target-arrow-fill":"filled","source-arrow-fill":"filled","curve-style":"bezier"}},{selector:"edge.none_to_none",style:{width:"4px","line-style":"dashed","target-arrow-shape":"triangle","target-arrow-fill":"filled","source-arrow-fill":"filled"}},{selector:"edge.all_to_all",style:{width:"4px","line-style":"solid"}},{selector:"edge.edge_unknown",style:{width:"4px","line-style":"dotted"}},{selector:"edge.total",style:{"source-arrow-shape":"circle","source-arrow-fill":"filled"}},{selector:"edge.functional",style:{"source-arrow-shape":"square"}},{selector:"edge.injective",style:{"target-arrow-shape":"triangle-backcurve"}},{selector:"edge.surjective",style:{"target-arrow-fill":"filled"}},{selector:"node:selected",style:{"overlay-opacity":0}},{selector:"edge:selected",style:{"overlay-opacity":0}},{selector:"edge.selected_edge",style:{width:"6px","line-color":"#007acc","target-arrow-color":"#007acc","source-arrow-color":"#007acc"}},{selector:"node.highlighted",style:{"overlay-color":"#007acc","overlay-opacity":.25,"overlay-padding":6}},{selector:"node.selected_node",style:{"background-color":"#999"}}],Fu=[{selector:"node",style:{content:"data(label)","text-wrap":"wrap","text-outline-width":"3px","text-outline-color":"#888","text-valign":"center",color:"#fff",width:"50",height:"50","border-color":"#000","background-color":"#888"}},{selector:"node.state",style:{}},{selector:"node.bottom_state",style:{"background-color":"#000","text-outline-color":"#000"}},{selector:"edge",style:{content:"data(label)",width:"4px","line-style":"solid","edge-text-rotation":"none","curve-style":"bezier",color:"#c8c8c8","text-outline-width":"2px","text-outline-color":"#1a1a2e","font-size":"11px"}},{selector:"edge.transition_join",style:{"target-arrow-shape":"triangle-backcurve"}},{selector:"edge.transition_action",style:{"target-arrow-shape":"triangle"}},{selector:"edge.cover",style:{content:"","target-arrow-shape":"triangle","line-style":"dashed"}},{selector:"node:selected",style:{"overlay-opacity":.2}},{selector:"edge:selected",style:{"overlay-opacity":.2}},{selector:"node.highlighted",style:{"overlay-color":"#007acc","overlay-opacity":.3,"overlay-padding":8}}],Ru=[{selector:"node",style:{content:"data(label)","text-wrap":"wrap","text-outline-width":"3px","text-outline-color":"#888","text-valign":"center",color:"#fff",width:"50",height:"50","border-color":"#000","background-color":"#888"}},{selector:"edge",style:{"target-arrow-shape":"triangle","curve-style":"bezier"}},{selector:"node.refuted",style:{"background-color":"#000","text-outline-color":"#000"}},{selector:"node:selected",style:{"overlay-opacity":.2}},{selector:"edge:selected",style:{"overlay-opacity":.2}}];function Mu(e){return e||globalThis.window}function Vu(e){return e&&e.cytoscape||globalThis.cytoscape}function Gu(e){return e&&e.cytoscapeDagre||globalThis.cytoscapeDagre}function Hu(e,t){return{...e,...t||{}}}class Ou{constructor(t,n,{win:i=globalThis.window}={}){this.containerId=t,this.win=Mu(i),this.doc=this.win&&this.win.document;const s=Vu(this.win);try{this.cy=s({container:this.doc&&this.doc.getElementById(t),style:n,layout:{name:"preset"},userZoomingEnabled:!0,userPanningEnabled:!0,boxSelectionEnabled:!1,minZoom:.2,maxZoom:5})}catch(r){const o=`FATAL: Cytoscape initialization failed for "${t}": ${r.message}
This usually means a stylesheet has an invalid data() mapper for a property that requires a concrete value.`;throw console.error(o,r),this.win&&typeof this.win.alert=="function"&&this.win.alert(o),r}this._cxttapOk=!1,this.cy.on("cxttap",()=>{this._cxttapOk=!0});const a=Gu(this.win);s&&a&&typeof s.use=="function"&&s.use(a),this._clickCallbacks={},this._rightClickCallbacks={}}update(t,n){if(this.cy.elements().remove(),!t||t.length===0)return;const i=t.map(s=>(s.group==="nodes"&&(s.data.width||(s.data.width=s.data.label?Math.max(50,s.data.label.length*8+20):50),s.data.height||(s.data.height=50),s.data.shape||(s.data.shape="ellipse"),s.data.border_color||(s.data.border_color="#000")),s));this.cy.add(i),this.cy.nodes().forEach(s=>{const a=s.data("border_color");a&&s.style("border-color",a)}),n?(Object.entries(n).forEach(([s,a])=>{const r=this.cy.getElementById(s);r.length>0&&r.position(a)}),this.cy.fit(void 0,30)):this.runLayout()}runLayout(t){const n=Hu({name:"dagre",rankDir:"TB",nodeSep:50,rankSep:80,edgeSep:10,animate:!1,fit:!0,padding:30},t);try{this.cy.layout(n).run()}catch(i){console.warn("Dagre layout failed, falling back to grid:",i),this.cy.layout({name:"grid",fit:!0,padding:30}).run()}}onNodeClick(t){this.cy.on("tap","node",n=>{t(n.target.data(),n)})}onNodeRightClick(t){this.cy.on("cxttap","node",n=>{const i=n.renderedPosition||n.target.renderedPosition();t(n.target.data(),i,n)})}onEdgeClick(t){this.cy.on("tap","edge",n=>{t(n.target.data(),n)})}onEdgeRightClick(t){this.cy.on("cxttap","edge",n=>{const i=n.renderedPosition||n.target.renderedMidpoint();t(n.target.data(),i,n)})}onBackgroundClick(t){this.cy.on("tap",n=>{n.target===this.cy&&t(n)})}getSelectedNodes(){const t=[];return this.cy.$("node:selected").forEach(n=>{t.push(n.data())}),t}highlightNode(t){this.cy.nodes().removeClass("highlighted");const n=this.cy.getElementById(t);n.length>0&&n.addClass("highlighted")}clearHighlights(){this.cy.nodes().removeClass("highlighted")}centerOnNode(t){const n=this.cy.getElementById(t);n.length>0&&this.cy.animate({center:{eles:n},duration:300})}fit(){this.cy.fit(void 0,30)}resize(){this.cy.resize(),this.cy.fit(void 0,30)}destroy(){this.cy&&(this.cy.destroy(),this.cy=null)}healthCheck(){const t=this.containerId;try{if(!this.cy)throw new Error("cy instance is null");this.cy.add({group:"nodes",data:{id:"__healthcheck__",label:"test",border_color:"#000",shape:"ellipse",width:10,height:10}});const n=this.cy.getElementById("__healthcheck__");if(n.length===0)throw new Error("test node not found after add");n.style("border-color"),n.remove();const i=()=>{};return this.cy.on("tap",i),this.cy.off("tap",i),this.win[`__ivyGraphHealthy_${t}`]=!0,!0}catch(n){const i=`IvyGraph health check FAILED for "${t}": ${n.message}`;return console.error(i,n),this.win&&typeof this.win.alert=="function"&&this.win.alert(i),this.win[`__ivyGraphHealthy_${t}`]=!1,!1}}}function ju(e=globalThis.window){e&&(e.IvyGraph||(e.IvyGraph=Ou),e.CONCEPT_STYLE||(e.CONCEPT_STYLE=Bu),e.ARG_STYLE||(e.ARG_STYLE=Fu),e.PROOF_STYLE||(e.PROOF_STYLE=Ru))}const $t="ivy_sessions",Cn="ivy_last_session",zt="ivy_sess_";function $u(e){return e||globalThis.window}function kt(e){return e&&e.__ivyVueBridge}function At(e){return e&&e.localStorage}function zu(e,t){if(!t)return"";let n="";t.api&&t.api.sessionId?n=e.getSessionIdFromURL()||t.api.sessionId:n=t.sessionId||"";const i=t._persistedFileName||t.fileName||"",s=t._persistedFilePath||t.filePath||i;return`${n}|${s}|${i}`}function Wt(e,t,n=null){const i=e&&e.getItem(t);return i?JSON.parse(i):n}function Wu(e=globalThis.window){const t=$u(e),n=t&&t.document,i={MAX_SESSIONS:1e3,HANDLE_DB:"ivy_file_handles",HANDLE_STORE:"handles",save(s){if(!s||!s.api||!s.api.sessionId)return;const a=At(t);if(!a)return;const r=i.getSessionIdFromURL()||s.api.sessionId;try{const o={sessionId:r,timestamp:Date.now(),fileName:s._persistedFileName||"",filePath:s._persistedFilePath||s._persistedFileName||"",fileContent:s._persistedFileContent||"",selectedArgNode:s.selectedArgNode||null,mode:i._getMode(),selectedConceptNodes:i._getSelectedConceptNodes(s),toggles:i._getToggles(),edgeVisibility:s._edgeVisibility||{},labelVisibility:s._labelVisibility||{},argElements:i._getCyElements(s.argGraph),conceptElements:i._getCyElements(s.conceptGraph),conceptRelations:s._persistedConceptRelations||null,analysisState:s.buildAnalysisState?s.buildAnalysisState():null};a.setItem(`${zt}${r}`,JSON.stringify(o)),a.setItem(Cn,r),i._updateSessionList(r)}catch(o){console.warn("IvyPersist.save failed:",o)}},load(){const s=At(t);if(!s)return null;try{const a=i.getSessionIdFromURL()||s.getItem(Cn);return a?Wt(s,`${zt}${a}`,null):null}catch(a){return console.warn("IvyPersist.load failed:",a),null}},loadSession(s){const a=At(t);if(!a||!s)return null;try{return Wt(a,`${zt}${s}`,null)}catch{return null}},deleteSession(s){const a=At(t);if(!(!a||!s))try{a.removeItem(`${zt}${s}`);const r=Wt(a,$t,[]).filter(o=>o!==s);a.setItem($t,JSON.stringify(r)),a.getItem(Cn)===s&&a.removeItem(Cn)}catch(r){console.warn("IvyPersist.deleteSession failed:",r)}},listSessions(){const s=At(t);if(!s)return[];try{return Wt(s,$t,[]).flatMap(a=>{const r=i.loadSession(a);return r?[{id:a,fileName:r.fileName||"(unnamed)",filePath:r.filePath||r.fileName||"",timestamp:r.timestamp||0}]:[]})}catch{return[]}},_handleKey(s){return zu(i,s)},_openHandleDB(){return new Promise((s,a)=>{if(!t||!t.indexedDB){a(new Error("IndexedDB unavailable"));return}const r=t.indexedDB.open(i.HANDLE_DB,1);r.onupgradeneeded=()=>{r.result.createObjectStore(i.HANDLE_STORE)},r.onsuccess=()=>s(r.result),r.onerror=()=>a(r.error||new Error("open IndexedDB failed"))})},async saveFileHandle(s){if(!(!s||!s._fileHandle))try{const a=i._handleKey(s);if(!a)return;let r="";s._persistedFileName&&s.api&&s.api.sessionId&&(r=`${i.getSessionIdFromURL()||s.api.sessionId}|${s._persistedFileName}|${s._persistedFileName}`);const o=await i._openHandleDB();await new Promise((l,d)=>{const c=o.transaction(i.HANDLE_STORE,"readwrite"),h=c.objectStore(i.HANDLE_STORE);h.put(s._fileHandle,a),r&&r!==a&&h.put(s._fileHandle,r),c.oncomplete=l,c.onerror=()=>d(c.error||new Error("store file handle failed"))}),o.close()}catch(a){console.warn("IvyPersist.saveFileHandle failed:",a)}},async loadFileHandle(s){try{const a=i._handleKey(s);if(!a)return null;const r=await i._openHandleDB(),o=[a];if(s.fileName){const h=`${s.sessionId||""}|${s.fileName}|${s.fileName}`;h&&h!==a&&o.push(h)}const l=s.fileName||"",d=s.filePath||l,c=await new Promise((h,m)=>{const B=r.transaction(i.HANDLE_STORE,"readonly").objectStore(i.HANDLE_STORE);let D=0;const W=()=>{if(!l){h(null);return}const V=B.openCursor();let H=null;V.onsuccess=()=>{const N=V.result;if(!N){h(H);return}const[Y,y="",P=""]=String(N.key||"").split("|");if(P===l&&(y===d||y===l)){h(N.value);return}!H&&P===l&&(H=N.value),N.continue()},V.onerror=()=>m(V.error||new Error("scan file handles failed"))},U=()=>{if(D>=o.length){W();return}const V=B.get(o[D]);D+=1,V.onsuccess=()=>{V.result?h(V.result):U()},V.onerror=()=>m(V.error||new Error("load file handle failed"))};U()});return r.close(),c||null}catch(a){return console.warn("IvyPersist.loadFileHandle failed:",a),null}},async restore(s,a){if(!a||!a.fileContent)return!1;s.controls.setStatus("Restoring session...");try{s._persistedFileName=a.fileName,s._persistedFilePath=a.filePath||a.fileName||"",s._fileHandle=await i.loadFileHandle(a);let r=a.fileContent||"";if(s._fileHandle)try{const c=await s._fileHandle.getFile();r=await c.text(),s._persistedFileName=c.name||s._persistedFileName,s._persistedFilePath=c.path||s._persistedFilePath||s._persistedFileName}catch(c){console.warn("IvyPersist.restore: could not read disk file handle, using cached content:",c)}s._persistedFileContent=r,s.setEditorContent&&s.setEditorContent(r),s._updateEditorLabel&&s._updateEditorLabel();let o=!0,l=null;try{const c=t.Blob||globalThis.Blob,h=t.File||globalThis.File,m=new c([r],{type:"text/plain"}),E=new h([m],s._persistedFileName||a.fileName||"restored.ivy");await s.api.loadFile(E)}catch(c){o=!1,console.warn("IvyPersist.restore: server rejected file (parse error):",c.message||String(c))}if(a.edgeVisibility&&(s._edgeVisibility=a.edgeVisibility),a.labelVisibility&&(s._labelVisibility=a.labelVisibility),o){const c=await s.api.getARG();c&&c.elements&&s.argGraph.update(c.elements,c.positions),l=await s.api.getConceptGraph(),l&&l.elements&&s.conceptGraph.update(l.elements,l.positions),s._persistedConceptRelations=l,s.populateStateCheckboxes(l)}a.mode&&i._setMode(a.mode),a.selectedConceptNodes&&s.conceptGraph&&s.conceptGraph.cy&&a.selectedConceptNodes.forEach(c=>{const h=s.conceptGraph.cy.getElementById(c);h.length>0&&h.addClass("selected_node")}),a.toggles&&i._setToggles(a.toggles);const d=i._buildVisibilityFromCheckboxes();return s._edgeVisibility=d.edges,s._labelVisibility=d.labels,s._applyEdgeVisibility(),s._applyNodeLabels(),a.selectedArgNode&&(s.selectedArgNode=a.selectedArgNode),a.analysisState&&s.loadAnalysisStateObject&&(a.analysisState.fileContent=r,a.analysisState.fileName=s._persistedFileName||a.fileName||a.analysisState.fileName,a.analysisState.filePath=s._persistedFilePath||a.filePath||a.analysisState.filePath,await s.loadAnalysisStateObject(a.analysisState)),i.setFileName(a.fileName,a.filePath),i.setSessionIdInURL(a.sessionId),s._fileHandle&&await i.saveFileHandle(s),s.controls.setStatus(`${o?"Restored":"Loaded (parse error)"}: ${a.fileName||"session"}`,o?"success":"warning"),!0}catch(r){return console.error("IvyPersist.restore failed:",r),s.controls.setStatus(`Restore failed: ${r.message}`,"error"),!1}},_updateSessionList(s){const a=At(t);if(!a)return;const r=Wt(a,$t,[]).filter(o=>o!==s);for(r.unshift(s);r.length>i.MAX_SESSIONS;){const o=r.pop();a.removeItem(`${zt}${o}`)}a.setItem($t,JSON.stringify(r))},_getMode(){const s=kt(t);if(s&&typeof s.getMode=="function")return s.getMode()||"pdr";const a=n&&n.getElementById("mode-select");return a?a.value:"pdr"},_setMode(s){const a=kt(t);if(a&&typeof a.setMode=="function"){a.setMode(s);return}const r=n&&n.getElementById("mode-select");r&&s&&(r.value=s)},_getSelectedConceptNodes(s){const a=[];return s.conceptGraph&&s.conceptGraph.cy&&s.conceptGraph.cy.nodes(".selected_node").forEach(r=>{a.push(r.id())}),a},_getToggles(){const s=kt(t);if(s&&typeof s.getStateRelationToggles=="function")return s.getStateRelationToggles()||{};const a={},r=n&&n.getElementById("state-checkbox-body");return r&&r.querySelectorAll('input[type="radio"], input[type="checkbox"]').forEach(o=>{o.name&&(a[`${o.name}|${o.value}`]=o.checked)}),a},_setToggles(s={}){const a=kt(t);if(a&&typeof a.setStateRelationToggles=="function"){a.setStateRelationToggles(s||{});return}const r=n&&n.getElementById("state-checkbox-body");r&&r.querySelectorAll('input[type="radio"], input[type="checkbox"]').forEach(o=>{const l=`${o.name}|${o.value}`;Object.prototype.hasOwnProperty.call(s,l)&&(o.checked=s[l])})},_getCyElements(s){return!s||!s.cy?null:s.cy.json().elements},_buildVisibilityFromCheckboxes(){const s=kt(t);if(s&&typeof s.buildStateRelationVisibility=="function")return s.buildStateRelationVisibility()||{edges:{},labels:{}};const a={},r={},o=n&&n.getElementById("state-checkbox-body");if(!o)return{edges:a,labels:r};const l=["all_to_all","edge_unknown","none_to_none","transitive"],d=["node_necessarily","node_maybe","node_necessarily_not"];return o.querySelectorAll("tr").forEach(c=>{const h=Array.from(c.querySelectorAll('input[type="checkbox"]')),m=c.querySelector(".name-col a"),E=m?m.textContent.trim():"";E&&(a[E]={},r[E]={},l.forEach((B,D)=>{h[D]&&(a[E][B]=h[D].checked)}),d.forEach((B,D)=>{h[D]&&(r[E][B]=h[D].checked)}))}),{edges:a,labels:r}},_buildEdgeVisibilityFromCheckboxes(){return i._buildVisibilityFromCheckboxes().edges},getSessionIdFromURL(){const s=t&&t.location&&t.location.hash;return s&&s.length>1?s.substring(1):null},setSessionIdInURL(s){s&&t&&t.history&&t.history.replaceState(null,"",`#${s}`)},setFileName(s,a){const r=kt(t);if(r&&typeof r.setLoadedFile=="function"){r.setLoadedFile(s||"",a||"");return}const o=n&&n.getElementById("loaded-file");if(o){const l=a||s||"";o.textContent=l,o.title=l}},truncatePath(s,a){if(!s||s.length<=a)return s||"";let r=s.lastIndexOf("/");r<0&&(r=s.lastIndexOf("\\"));const o=r>=0?s.substring(0,r):"";return o?o.length<=a?o:`...${o.substring(o.length-a+3)}`:""}};return i}class Uu extends Za{constructor(t=""){const n=new Ja({baseURL:t});super(new Ps({client:n})),this.baseURL=t||""}}class Ku{constructor(t){this.api=t,this.edgeToggles={},this.labelToggles={},this._contextMenuVisible=!1}buildEdgeToggles(){this.edgeToggles={}}buildLabelToggles(){this.labelToggles={}}getEdgeToggleState(){return{}}getLabelToggleState(){return{}}showContextMenu(t,n,i=[]){const s=globalThis.window&&globalThis.window.__ivyVueBridge;s&&typeof s.showContextMenu=="function"&&(s.showContextMenu(t,n,i),this._contextMenuVisible=!0)}hideContextMenu(){const t=globalThis.window&&globalThis.window.__ivyVueBridge;t&&typeof t.hideContextMenu=="function"&&t.hideContextMenu(),this._contextMenuVisible=!1}isContextMenuVisible(){return this._contextMenuVisible}showInfo(t,n){const i=globalThis.window&&globalThis.window.__ivyVueBridge;i&&typeof i.updateDetails=="function"&&i.updateDetails({shortInfo:t,longInfo:n})}clearInfo(){const t=globalThis.window&&globalThis.window.__ivyVueBridge;t&&typeof t.clearDetails=="function"&&t.clearDetails()}setStatus(t,n=""){const i=globalThis.window&&globalThis.window.__ivyVueBridge;i&&typeof i.updateStatus=="function"&&i.updateStatus(t,n||"")}showLoading(t="Loading..."){const n=globalThis.window&&globalThis.window.__ivyVueBridge;n&&typeof n.showLoading=="function"&&n.showLoading(t||"Loading...")}hideLoading(){const t=globalThis.window&&globalThis.window.__ivyVueBridge;t&&typeof t.hideLoading=="function"&&t.hideLoading()}}function qu(e=globalThis.window){e&&(ju(e),e.IvyAPI||(e.IvyAPI=Uu),e.IvyControls||(e.IvyControls=Ku),e.IvyPersist||(e.IvyPersist=Wu(e)))}const Yu=[];function Xu(e,t){return!!(e&&e.querySelector(`script[data-ivy-legacy-script="${t}"]`))}function Ju(e,t){return Xu(e,t)?Promise.resolve():new Promise((n,i)=>{const s=e.createElement("script");s.src=t,s.async=!1,s.dataset.ivyLegacyScript=t,s.onload=()=>n(),s.onerror=()=>i(new Error(`Failed to load legacy Ivy script: ${t}`)),e.body.appendChild(s)})}async function Zu({doc:e=globalThis.document,win:t=globalThis.window,scripts:n=Yu,installAppRuntime:i=Nu}={}){if(!(!e||!t)&&(t.__IVY_VUE_OWNS_BOOT__=!0,qu(t),typeof t.startIvyApp!="function"&&i({win:t}),typeof t.startIvyApp!="function"))for(const s of n)await Ju(e,s)}async function Qu(){await Zu();const e=Ml(yu),t=Hl();e.use(t),Lu({app:e,pinia:t}),e.mount("#ivy-vue-root")}Qu().catch(e=>{window.__ivyInitError=e&&e.message?e.message:String(e),console.error("Failed to boot Ivy Vue app:",e)});
//# sourceMappingURL=ivyweb.js.map
