import{_ as U}from"./dpm.274c7743.js";import{i as V}from"./icon.43ea7822.js";import{_ as B,r as c,o as D,b as l,c as L,e as a,d as t,f as s,F as N,i as z,C as F,j as h,k as o,m as M,g as n,y as v,u as R,h as q,p as G,q as P}from"./index.4e24e414.js";const f=r=>(G("data-v-eb79cae9"),r=r(),P(),r),T={class:"pj-top"},A={class:"pj-title"},H=f(()=>a("img",{src:U,style:{width:"25px",height:"25px","padding-right":"10px"}},null,-1)),J=f(()=>a("div",{style:{"padding-right":"10px"}},"\u7B14\u8BB0\u7CFB\u7EDF",-1)),K={style:{width:"40px",height:"40px","border-radius":"25px",overflow:"hidden"}},O={__name:"Index",setup(r){const d=q(),_=z(),g=F(),u=c(""),i=c({});i.value=_.getters.getUserInfo;const x=()=>{l.get("/api/system/version").then(e=>{u.value=e.data}).catch(e=>{h.error({message:e.response.data,duration:2e3,showClose:!0})})},y=()=>{l.delete("/api/logout").then(e=>{d.push({path:"/"})}).catch(e=>{h.error({message:e.response.data,duration:2e3,showClose:!0})})},I=()=>{d.push({name:"User"})};document.addEventListener("drop",function(e){e.preventDefault()},!1),document.addEventListener("dragover",function(e){e.preventDefault()},!1);const p=c("");return D(()=>{l.get("/api/check").then(e=>{e.data.imgSrc="/api/user/avatar?id="+e.data.id,_.commit("saveUserInfo",e.data),i.value=e.data}).catch(()=>{}),x(),p.value=g.fullPath}),(e,Q)=>{const k=o("el-tag"),m=o("el-menu-item"),w=o("el-avatar"),b=o("el-image"),C=o("el-button"),S=o("el-link"),j=o("el-menu"),E=o("router-view");return M(),L(N,null,[a("div",T,[t(j,{class:"pj-menu",mode:"horizontal",router:"","default-active":p.value},{default:s(()=>[a("div",A,[H,J,t(k,{type:"info",style:{"font-size":"18px"}},{default:s(()=>[n(v(u.value.systemVersion),1)]),_:1})]),t(m,{index:"/index/noteList"},{default:s(()=>[n("\u7B14\u8BB0\u5217\u8868")]),_:1}),t(m,{index:"/Index/userGroup"},{default:s(()=>[n("\u7528\u6237\u7EC4\u7BA1\u7406")]),_:1}),a("div",{class:"i-avatar",onClick:I},[a("div",K,[t(b,{src:i.value.imgSrc,style:{width:"40px",height:"40px"}},{error:s(()=>[t(w,{src:R(V)},null,8,["src"])]),_:1},8,["src"])]),t(C,{link:"",class:"i-name"},{default:s(()=>[n(v(i.value.name),1)]),_:1})]),t(S,{type:"info",style:{"margin-top":"8px","margin-right":"10px"},onClick:y},{default:s(()=>[n("\u9000\u51FA\u767B\u5F55")]),_:1})]),_:1},8,["default-active"])]),t(E)],64)}}},Z=B(O,[["__scopeId","data-v-eb79cae9"]]);export{Z as default};