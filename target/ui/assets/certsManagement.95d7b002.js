import{_ as U,r as d,o as E,w as I,c as N,e as p,d as t,f as o,b as f,j as _,k as a,l as P,m as T,g as u,u as z,aP as R,aO as j}from"./index.4e24e414.js";const F={class:"um-body","element-loading-text":"\u6570\u636E\u52A0\u8F7D\u4E2D..."},O={class:"um-operation"},q={class:"dialog-footer"},A={__name:"certsManagement",setup(G){const r=d(!1),v=d([]),c=d(!1),s=d(null),C=()=>{r.value=!1,s.value.value=null},m=()=>{c.value=!0;var l="/api/rootCerts/list";f.get(l).then(e=>{v.value=e.data}).catch(e=>{_.error({message:e.response.data,duration:2e3,showClose:!0})}).finally(()=>{c.value=!1})},w=()=>{let l=new FormData;console.log(s.value.files[0]),s.value.files[0]!==void 0&&l.append("files",s.value.files[0]),r.value=!1,c.value=!0,f.post("/api/rootCerts/upload",l).then(e=>{_.success({message:"\u4E0A\u4F20\u6210\u529F",duration:1e3,showClose:!0}),s.value.value=null,m()}).catch(e=>{_.error({message:e.response.data,duration:1e3,showClose:!0})}).finally(()=>{c.value=!1})},b=l=>{window.location.href=`/api/rootCerts/download?name=${encodeURIComponent(l)}`},y=l=>{j.confirm("\u786E\u8BA4\u662F\u5426\u5220\u9664 "+l+" \u8BC1\u4E66\uFF1F","\u786E\u8BA4\u4FE1\u606F",{confirmButtonText:"\u786E\u5B9A",cancelButtonText:"\u53D6\u6D88",type:"warning"}).then(()=>{f.delete(`/api/rootCerts/remove?name=${encodeURIComponent(l)}`).then(e=>{_.success({message:"\u5220\u9664\u6210\u529F",duration:2e3,showClose:!0}),m()})}).catch(()=>{})};return E(()=>{m()}),(l,e)=>{const h=a("Plus"),x=a("el-icon"),i=a("el-button"),g=a("el-table-column"),k=a("el-table"),V=a("el-form-item"),$=a("el-form"),B=a("el-dialog"),D=P("loading");return I((T(),N("div",F,[p("div",O,[t(i,{class:"um-btn",type:"primary",onClick:e[0]||(e[0]=n=>r.value=!0)},{default:o(()=>[t(x,null,{default:o(()=>[t(h)]),_:1}),u(" \xA0\u4E0A\u4F20\u6839\u8BC1\u4E66 ")]),_:1})]),t(k,{data:v.value,"empty-text":"\u6682\u65E0\u6839\u8BC1\u4E66",style:{width:"100%"},border:""},{default:o(()=>[t(g,{label:"\u8BC1\u4E66\u540D\u79F0",prop:"name"}),t(g,{label:"\u64CD\u4F5C\u680F"},{default:o(n=>[t(i,{type:"primary",size:"small",class:"um-btn",onClick:M=>{b(n.row.name)}},{default:o(()=>[u("\u4E0B\u8F7D")]),_:2},1032,["onClick"]),t(i,{size:"small",type:"danger",class:"um-btn",icon:z(R),onClick:M=>y(n.row.name)},{default:o(()=>[u("\u5220\u9664")]),_:2},1032,["icon","onClick"])]),_:1})]),_:1},8,["data"]),t(B,{modelValue:r.value,"onUpdate:modelValue":e[3]||(e[3]=n=>r.value=n),"close-on-click-modal":!1,"show-close":!1,title:"\u4E0A\u4F20\u6839\u8BC1\u4E66"},{footer:o(()=>[p("span",q,[t(i,{onClick:e[1]||(e[1]=n=>C())},{default:o(()=>[u("\u53D6\u6D88")]),_:1}),t(i,{type:"primary",onClick:e[2]||(e[2]=n=>w())},{default:o(()=>[u("\u63D0\u4EA4")]),_:1})])]),default:o(()=>[t($,{"label-width":"120px"},{default:o(()=>[t(V,{label:"\u6587\u4EF6",style:{width:"80%"}},{default:o(()=>[p("input",{type:"file",placeholder:"\u4E0A\u4F20\u6587\u4EF6",ref_key:"file",ref:s},null,512)]),_:1})]),_:1})]),_:1},8,["modelValue"])])),[[D,c.value]])}}},J=U(A,[["__scopeId","data-v-87d9e1a6"]]);export{J as default};