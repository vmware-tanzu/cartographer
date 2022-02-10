import{f as x,R as Z,q as l,M as T,U as K}from"./vendor.3a37258e.js";var Fe=2*60*1e3,je=function(){function n(t){var a=this;this._defaults=t,this._worker=null,this._idleCheckInterval=window.setInterval(function(){return a._checkIfIdle()},30*1e3),this._lastUsedTime=0,this._configChangeListener=this._defaults.onDidChange(function(){return a._stopWorker()})}return n.prototype._stopWorker=function(){this._worker&&(this._worker.dispose(),this._worker=null),this._client=null},n.prototype.dispose=function(){clearInterval(this._idleCheckInterval),this._configChangeListener.dispose(),this._stopWorker()},n.prototype._checkIfIdle=function(){if(!!this._worker){var t=Date.now()-this._lastUsedTime;t>Fe&&this._stopWorker()}},n.prototype._getClient=function(){return this._lastUsedTime=Date.now(),this._client||(this._worker=x.createWebWorker({moduleId:"vs/language/css/cssWorker",label:this._defaults.languageId,createData:{options:this._defaults.options,languageId:this._defaults.languageId}}),this._client=this._worker.getProxy()),this._client},n.prototype.getLanguageServiceWorker=function(){for(var t=this,a=[],r=0;r<arguments.length;r++)a[r]=arguments[r];var e;return this._getClient().then(function(i){e=i}).then(function(i){return t._worker.withSyncedResources(a)}).then(function(i){return e})},n}(),ee;(function(n){n.MIN_VALUE=-2147483648,n.MAX_VALUE=2147483647})(ee||(ee={}));var W;(function(n){n.MIN_VALUE=0,n.MAX_VALUE=2147483647})(W||(W={}));var k;(function(n){function t(r,e){return r===Number.MAX_VALUE&&(r=W.MAX_VALUE),e===Number.MAX_VALUE&&(e=W.MAX_VALUE),{line:r,character:e}}n.create=t;function a(r){var e=r;return u.objectLiteral(e)&&u.uinteger(e.line)&&u.uinteger(e.character)}n.is=a})(k||(k={}));var m;(function(n){function t(r,e,i,o){if(u.uinteger(r)&&u.uinteger(e)&&u.uinteger(i)&&u.uinteger(o))return{start:k.create(r,e),end:k.create(i,o)};if(k.is(r)&&k.is(e))return{start:r,end:e};throw new Error("Range#create called with invalid arguments["+r+", "+e+", "+i+", "+o+"]")}n.create=t;function a(r){var e=r;return u.objectLiteral(e)&&k.is(e.start)&&k.is(e.end)}n.is=a})(m||(m={}));var X;(function(n){function t(r,e){return{uri:r,range:e}}n.create=t;function a(r){var e=r;return u.defined(e)&&m.is(e.range)&&(u.string(e.uri)||u.undefined(e.uri))}n.is=a})(X||(X={}));var ne;(function(n){function t(r,e,i,o){return{targetUri:r,targetRange:e,targetSelectionRange:i,originSelectionRange:o}}n.create=t;function a(r){var e=r;return u.defined(e)&&m.is(e.targetRange)&&u.string(e.targetUri)&&(m.is(e.targetSelectionRange)||u.undefined(e.targetSelectionRange))&&(m.is(e.originSelectionRange)||u.undefined(e.originSelectionRange))}n.is=a})(ne||(ne={}));var $;(function(n){function t(r,e,i,o){return{red:r,green:e,blue:i,alpha:o}}n.create=t;function a(r){var e=r;return u.numberRange(e.red,0,1)&&u.numberRange(e.green,0,1)&&u.numberRange(e.blue,0,1)&&u.numberRange(e.alpha,0,1)}n.is=a})($||($={}));var te;(function(n){function t(r,e){return{range:r,color:e}}n.create=t;function a(r){var e=r;return m.is(e.range)&&$.is(e.color)}n.is=a})(te||(te={}));var re;(function(n){function t(r,e,i){return{label:r,textEdit:e,additionalTextEdits:i}}n.create=t;function a(r){var e=r;return u.string(e.label)&&(u.undefined(e.textEdit)||E.is(e))&&(u.undefined(e.additionalTextEdits)||u.typedArray(e.additionalTextEdits,E.is))}n.is=a})(re||(re={}));var D;(function(n){n.Comment="comment",n.Imports="imports",n.Region="region"})(D||(D={}));var ie;(function(n){function t(r,e,i,o,s){var c={startLine:r,endLine:e};return u.defined(i)&&(c.startCharacter=i),u.defined(o)&&(c.endCharacter=o),u.defined(s)&&(c.kind=s),c}n.create=t;function a(r){var e=r;return u.uinteger(e.startLine)&&u.uinteger(e.startLine)&&(u.undefined(e.startCharacter)||u.uinteger(e.startCharacter))&&(u.undefined(e.endCharacter)||u.uinteger(e.endCharacter))&&(u.undefined(e.kind)||u.string(e.kind))}n.is=a})(ie||(ie={}));var B;(function(n){function t(r,e){return{location:r,message:e}}n.create=t;function a(r){var e=r;return u.defined(e)&&X.is(e.location)&&u.string(e.message)}n.is=a})(B||(B={}));var I;(function(n){n.Error=1,n.Warning=2,n.Information=3,n.Hint=4})(I||(I={}));var ae;(function(n){n.Unnecessary=1,n.Deprecated=2})(ae||(ae={}));var oe;(function(n){function t(a){var r=a;return r!=null&&u.string(r.href)}n.is=t})(oe||(oe={}));var U;(function(n){function t(r,e,i,o,s,c){var g={range:r,message:e};return u.defined(i)&&(g.severity=i),u.defined(o)&&(g.code=o),u.defined(s)&&(g.source=s),u.defined(c)&&(g.relatedInformation=c),g}n.create=t;function a(r){var e,i=r;return u.defined(i)&&m.is(i.range)&&u.string(i.message)&&(u.number(i.severity)||u.undefined(i.severity))&&(u.integer(i.code)||u.string(i.code)||u.undefined(i.code))&&(u.undefined(i.codeDescription)||u.string((e=i.codeDescription)===null||e===void 0?void 0:e.href))&&(u.string(i.source)||u.undefined(i.source))&&(u.undefined(i.relatedInformation)||u.typedArray(i.relatedInformation,B.is))}n.is=a})(U||(U={}));var M;(function(n){function t(r,e){for(var i=[],o=2;o<arguments.length;o++)i[o-2]=arguments[o];var s={title:r,command:e};return u.defined(i)&&i.length>0&&(s.arguments=i),s}n.create=t;function a(r){var e=r;return u.defined(e)&&u.string(e.title)&&u.string(e.command)}n.is=a})(M||(M={}));var E;(function(n){function t(i,o){return{range:i,newText:o}}n.replace=t;function a(i,o){return{range:{start:i,end:i},newText:o}}n.insert=a;function r(i){return{range:i,newText:""}}n.del=r;function e(i){var o=i;return u.objectLiteral(o)&&u.string(o.newText)&&m.is(o.range)}n.is=e})(E||(E={}));var R;(function(n){function t(r,e,i){var o={label:r};return e!==void 0&&(o.needsConfirmation=e),i!==void 0&&(o.description=i),o}n.create=t;function a(r){var e=r;return e!==void 0&&u.objectLiteral(e)&&u.string(e.label)&&(u.boolean(e.needsConfirmation)||e.needsConfirmation===void 0)&&(u.string(e.description)||e.description===void 0)}n.is=a})(R||(R={}));var _;(function(n){function t(a){var r=a;return typeof r=="string"}n.is=t})(_||(_={}));var A;(function(n){function t(i,o,s){return{range:i,newText:o,annotationId:s}}n.replace=t;function a(i,o,s){return{range:{start:i,end:i},newText:o,annotationId:s}}n.insert=a;function r(i,o){return{range:i,newText:"",annotationId:o}}n.del=r;function e(i){var o=i;return E.is(o)&&(R.is(o.annotationId)||_.is(o.annotationId))}n.is=e})(A||(A={}));var H;(function(n){function t(r,e){return{textDocument:r,edits:e}}n.create=t;function a(r){var e=r;return u.defined(e)&&O.is(e.textDocument)&&Array.isArray(e.edits)}n.is=a})(H||(H={}));var P;(function(n){function t(r,e,i){var o={kind:"create",uri:r};return e!==void 0&&(e.overwrite!==void 0||e.ignoreIfExists!==void 0)&&(o.options=e),i!==void 0&&(o.annotationId=i),o}n.create=t;function a(r){var e=r;return e&&e.kind==="create"&&u.string(e.uri)&&(e.options===void 0||(e.options.overwrite===void 0||u.boolean(e.options.overwrite))&&(e.options.ignoreIfExists===void 0||u.boolean(e.options.ignoreIfExists)))&&(e.annotationId===void 0||_.is(e.annotationId))}n.is=a})(P||(P={}));var L;(function(n){function t(r,e,i,o){var s={kind:"rename",oldUri:r,newUri:e};return i!==void 0&&(i.overwrite!==void 0||i.ignoreIfExists!==void 0)&&(s.options=i),o!==void 0&&(s.annotationId=o),s}n.create=t;function a(r){var e=r;return e&&e.kind==="rename"&&u.string(e.oldUri)&&u.string(e.newUri)&&(e.options===void 0||(e.options.overwrite===void 0||u.boolean(e.options.overwrite))&&(e.options.ignoreIfExists===void 0||u.boolean(e.options.ignoreIfExists)))&&(e.annotationId===void 0||_.is(e.annotationId))}n.is=a})(L||(L={}));var S;(function(n){function t(r,e,i){var o={kind:"delete",uri:r};return e!==void 0&&(e.recursive!==void 0||e.ignoreIfNotExists!==void 0)&&(o.options=e),i!==void 0&&(o.annotationId=i),o}n.create=t;function a(r){var e=r;return e&&e.kind==="delete"&&u.string(e.uri)&&(e.options===void 0||(e.options.recursive===void 0||u.boolean(e.options.recursive))&&(e.options.ignoreIfNotExists===void 0||u.boolean(e.options.ignoreIfNotExists)))&&(e.annotationId===void 0||_.is(e.annotationId))}n.is=a})(S||(S={}));var q;(function(n){function t(a){var r=a;return r&&(r.changes!==void 0||r.documentChanges!==void 0)&&(r.documentChanges===void 0||r.documentChanges.every(function(e){return u.string(e.kind)?P.is(e)||L.is(e)||S.is(e):H.is(e)}))}n.is=t})(q||(q={}));var V=function(){function n(t,a){this.edits=t,this.changeAnnotations=a}return n.prototype.insert=function(t,a,r){var e,i;if(r===void 0?e=E.insert(t,a):_.is(r)?(i=r,e=A.insert(t,a,r)):(this.assertChangeAnnotations(this.changeAnnotations),i=this.changeAnnotations.manage(r),e=A.insert(t,a,i)),this.edits.push(e),i!==void 0)return i},n.prototype.replace=function(t,a,r){var e,i;if(r===void 0?e=E.replace(t,a):_.is(r)?(i=r,e=A.replace(t,a,r)):(this.assertChangeAnnotations(this.changeAnnotations),i=this.changeAnnotations.manage(r),e=A.replace(t,a,i)),this.edits.push(e),i!==void 0)return i},n.prototype.delete=function(t,a){var r,e;if(a===void 0?r=E.del(t):_.is(a)?(e=a,r=A.del(t,a)):(this.assertChangeAnnotations(this.changeAnnotations),e=this.changeAnnotations.manage(a),r=A.del(t,e)),this.edits.push(r),e!==void 0)return e},n.prototype.add=function(t){this.edits.push(t)},n.prototype.all=function(){return this.edits},n.prototype.clear=function(){this.edits.splice(0,this.edits.length)},n.prototype.assertChangeAnnotations=function(t){if(t===void 0)throw new Error("Text edit change is not configured to manage change annotations.")},n}(),ue=function(){function n(t){this._annotations=t===void 0?Object.create(null):t,this._counter=0,this._size=0}return n.prototype.all=function(){return this._annotations},Object.defineProperty(n.prototype,"size",{get:function(){return this._size},enumerable:!1,configurable:!0}),n.prototype.manage=function(t,a){var r;if(_.is(t)?r=t:(r=this.nextId(),a=t),this._annotations[r]!==void 0)throw new Error("Id "+r+" is already in use.");if(a===void 0)throw new Error("No annotation provided for id "+r);return this._annotations[r]=a,this._size++,r},n.prototype.nextId=function(){return this._counter++,this._counter.toString()},n}();(function(){function n(t){var a=this;this._textEditChanges=Object.create(null),t!==void 0?(this._workspaceEdit=t,t.documentChanges?(this._changeAnnotations=new ue(t.changeAnnotations),t.changeAnnotations=this._changeAnnotations.all(),t.documentChanges.forEach(function(r){if(H.is(r)){var e=new V(r.edits,a._changeAnnotations);a._textEditChanges[r.textDocument.uri]=e}})):t.changes&&Object.keys(t.changes).forEach(function(r){var e=new V(t.changes[r]);a._textEditChanges[r]=e})):this._workspaceEdit={}}return Object.defineProperty(n.prototype,"edit",{get:function(){return this.initDocumentChanges(),this._changeAnnotations!==void 0&&(this._changeAnnotations.size===0?this._workspaceEdit.changeAnnotations=void 0:this._workspaceEdit.changeAnnotations=this._changeAnnotations.all()),this._workspaceEdit},enumerable:!1,configurable:!0}),n.prototype.getTextEditChange=function(t){if(O.is(t)){if(this.initDocumentChanges(),this._workspaceEdit.documentChanges===void 0)throw new Error("Workspace edit is not configured for document changes.");var a={uri:t.uri,version:t.version},r=this._textEditChanges[a.uri];if(!r){var e=[],i={textDocument:a,edits:e};this._workspaceEdit.documentChanges.push(i),r=new V(e,this._changeAnnotations),this._textEditChanges[a.uri]=r}return r}else{if(this.initChanges(),this._workspaceEdit.changes===void 0)throw new Error("Workspace edit is not configured for normal text edit changes.");var r=this._textEditChanges[t];if(!r){var e=[];this._workspaceEdit.changes[t]=e,r=new V(e),this._textEditChanges[t]=r}return r}},n.prototype.initDocumentChanges=function(){this._workspaceEdit.documentChanges===void 0&&this._workspaceEdit.changes===void 0&&(this._changeAnnotations=new ue,this._workspaceEdit.documentChanges=[],this._workspaceEdit.changeAnnotations=this._changeAnnotations.all())},n.prototype.initChanges=function(){this._workspaceEdit.documentChanges===void 0&&this._workspaceEdit.changes===void 0&&(this._workspaceEdit.changes=Object.create(null))},n.prototype.createFile=function(t,a,r){if(this.initDocumentChanges(),this._workspaceEdit.documentChanges===void 0)throw new Error("Workspace edit is not configured for document changes.");var e;R.is(a)||_.is(a)?e=a:r=a;var i,o;if(e===void 0?i=P.create(t,r):(o=_.is(e)?e:this._changeAnnotations.manage(e),i=P.create(t,r,o)),this._workspaceEdit.documentChanges.push(i),o!==void 0)return o},n.prototype.renameFile=function(t,a,r,e){if(this.initDocumentChanges(),this._workspaceEdit.documentChanges===void 0)throw new Error("Workspace edit is not configured for document changes.");var i;R.is(r)||_.is(r)?i=r:e=r;var o,s;if(i===void 0?o=L.create(t,a,e):(s=_.is(i)?i:this._changeAnnotations.manage(i),o=L.create(t,a,e,s)),this._workspaceEdit.documentChanges.push(o),s!==void 0)return s},n.prototype.deleteFile=function(t,a,r){if(this.initDocumentChanges(),this._workspaceEdit.documentChanges===void 0)throw new Error("Workspace edit is not configured for document changes.");var e;R.is(a)||_.is(a)?e=a:r=a;var i,o;if(e===void 0?i=S.create(t,r):(o=_.is(e)?e:this._changeAnnotations.manage(e),i=S.create(t,r,o)),this._workspaceEdit.documentChanges.push(i),o!==void 0)return o},n})();var se;(function(n){function t(r){return{uri:r}}n.create=t;function a(r){var e=r;return u.defined(e)&&u.string(e.uri)}n.is=a})(se||(se={}));var ce;(function(n){function t(r,e){return{uri:r,version:e}}n.create=t;function a(r){var e=r;return u.defined(e)&&u.string(e.uri)&&u.integer(e.version)}n.is=a})(ce||(ce={}));var O;(function(n){function t(r,e){return{uri:r,version:e}}n.create=t;function a(r){var e=r;return u.defined(e)&&u.string(e.uri)&&(e.version===null||u.integer(e.version))}n.is=a})(O||(O={}));var de;(function(n){function t(r,e,i,o){return{uri:r,languageId:e,version:i,text:o}}n.create=t;function a(r){var e=r;return u.defined(e)&&u.string(e.uri)&&u.string(e.languageId)&&u.integer(e.version)&&u.string(e.text)}n.is=a})(de||(de={}));var F;(function(n){n.PlainText="plaintext",n.Markdown="markdown"})(F||(F={}));(function(n){function t(a){var r=a;return r===n.PlainText||r===n.Markdown}n.is=t})(F||(F={}));var Q;(function(n){function t(a){var r=a;return u.objectLiteral(a)&&F.is(r.kind)&&u.string(r.value)}n.is=t})(Q||(Q={}));var h;(function(n){n.Text=1,n.Method=2,n.Function=3,n.Constructor=4,n.Field=5,n.Variable=6,n.Class=7,n.Interface=8,n.Module=9,n.Property=10,n.Unit=11,n.Value=12,n.Enum=13,n.Keyword=14,n.Snippet=15,n.Color=16,n.File=17,n.Reference=18,n.Folder=19,n.EnumMember=20,n.Constant=21,n.Struct=22,n.Event=23,n.Operator=24,n.TypeParameter=25})(h||(h={}));var G;(function(n){n.PlainText=1,n.Snippet=2})(G||(G={}));var fe;(function(n){n.Deprecated=1})(fe||(fe={}));var ge;(function(n){function t(r,e,i){return{newText:r,insert:e,replace:i}}n.create=t;function a(r){var e=r;return e&&u.string(e.newText)&&m.is(e.insert)&&m.is(e.replace)}n.is=a})(ge||(ge={}));var le;(function(n){n.asIs=1,n.adjustIndentation=2})(le||(le={}));var he;(function(n){function t(a){return{label:a}}n.create=t})(he||(he={}));var ve;(function(n){function t(a,r){return{items:a||[],isIncomplete:!!r}}n.create=t})(ve||(ve={}));var z;(function(n){function t(r){return r.replace(/[\\`*_{}[\]()#+\-.!]/g,"\\$&")}n.fromPlainText=t;function a(r){var e=r;return u.string(e)||u.objectLiteral(e)&&u.string(e.language)&&u.string(e.value)}n.is=a})(z||(z={}));var pe;(function(n){function t(a){var r=a;return!!r&&u.objectLiteral(r)&&(Q.is(r.contents)||z.is(r.contents)||u.typedArray(r.contents,z.is))&&(a.range===void 0||m.is(a.range))}n.is=t})(pe||(pe={}));var me;(function(n){function t(a,r){return r?{label:a,documentation:r}:{label:a}}n.create=t})(me||(me={}));var _e;(function(n){function t(a,r){for(var e=[],i=2;i<arguments.length;i++)e[i-2]=arguments[i];var o={label:a};return u.defined(r)&&(o.documentation=r),u.defined(e)?o.parameters=e:o.parameters=[],o}n.create=t})(_e||(_e={}));var j;(function(n){n.Text=1,n.Read=2,n.Write=3})(j||(j={}));var we;(function(n){function t(a,r){var e={range:a};return u.number(r)&&(e.kind=r),e}n.create=t})(we||(we={}));var v;(function(n){n.File=1,n.Module=2,n.Namespace=3,n.Package=4,n.Class=5,n.Method=6,n.Property=7,n.Field=8,n.Constructor=9,n.Enum=10,n.Interface=11,n.Function=12,n.Variable=13,n.Constant=14,n.String=15,n.Number=16,n.Boolean=17,n.Array=18,n.Object=19,n.Key=20,n.Null=21,n.EnumMember=22,n.Struct=23,n.Event=24,n.Operator=25,n.TypeParameter=26})(v||(v={}));var ke;(function(n){n.Deprecated=1})(ke||(ke={}));var be;(function(n){function t(a,r,e,i,o){var s={name:a,kind:r,location:{uri:i,range:e}};return o&&(s.containerName=o),s}n.create=t})(be||(be={}));var xe;(function(n){function t(r,e,i,o,s,c){var g={name:r,detail:e,kind:i,range:o,selectionRange:s};return c!==void 0&&(g.children=c),g}n.create=t;function a(r){var e=r;return e&&u.string(e.name)&&u.number(e.kind)&&m.is(e.range)&&m.is(e.selectionRange)&&(e.detail===void 0||u.string(e.detail))&&(e.deprecated===void 0||u.boolean(e.deprecated))&&(e.children===void 0||Array.isArray(e.children))&&(e.tags===void 0||Array.isArray(e.tags))}n.is=a})(xe||(xe={}));var Ee;(function(n){n.Empty="",n.QuickFix="quickfix",n.Refactor="refactor",n.RefactorExtract="refactor.extract",n.RefactorInline="refactor.inline",n.RefactorRewrite="refactor.rewrite",n.Source="source",n.SourceOrganizeImports="source.organizeImports",n.SourceFixAll="source.fixAll"})(Ee||(Ee={}));var Ae;(function(n){function t(r,e){var i={diagnostics:r};return e!=null&&(i.only=e),i}n.create=t;function a(r){var e=r;return u.defined(e)&&u.typedArray(e.diagnostics,U.is)&&(e.only===void 0||u.typedArray(e.only,u.string))}n.is=a})(Ae||(Ae={}));var Ce;(function(n){function t(r,e,i){var o={title:r},s=!0;return typeof e=="string"?(s=!1,o.kind=e):M.is(e)?o.command=e:o.edit=e,s&&i!==void 0&&(o.kind=i),o}n.create=t;function a(r){var e=r;return e&&u.string(e.title)&&(e.diagnostics===void 0||u.typedArray(e.diagnostics,U.is))&&(e.kind===void 0||u.string(e.kind))&&(e.edit!==void 0||e.command!==void 0)&&(e.command===void 0||M.is(e.command))&&(e.isPreferred===void 0||u.boolean(e.isPreferred))&&(e.edit===void 0||q.is(e.edit))}n.is=a})(Ce||(Ce={}));var ye;(function(n){function t(r,e){var i={range:r};return u.defined(e)&&(i.data=e),i}n.create=t;function a(r){var e=r;return u.defined(e)&&m.is(e.range)&&(u.undefined(e.command)||M.is(e.command))}n.is=a})(ye||(ye={}));var Ie;(function(n){function t(r,e){return{tabSize:r,insertSpaces:e}}n.create=t;function a(r){var e=r;return u.defined(e)&&u.uinteger(e.tabSize)&&u.boolean(e.insertSpaces)}n.is=a})(Ie||(Ie={}));var Re;(function(n){function t(r,e,i){return{range:r,target:e,data:i}}n.create=t;function a(r){var e=r;return u.defined(e)&&m.is(e.range)&&(u.undefined(e.target)||u.string(e.target))}n.is=a})(Re||(Re={}));var Te;(function(n){function t(r,e){return{range:r,parent:e}}n.create=t;function a(r){var e=r;return e!==void 0&&m.is(e.range)&&(e.parent===void 0||n.is(e.parent))}n.is=a})(Te||(Te={}));var De;(function(n){function t(i,o,s,c){return new Ne(i,o,s,c)}n.create=t;function a(i){var o=i;return!!(u.defined(o)&&u.string(o.uri)&&(u.undefined(o.languageId)||u.string(o.languageId))&&u.uinteger(o.lineCount)&&u.func(o.getText)&&u.func(o.positionAt)&&u.func(o.offsetAt))}n.is=a;function r(i,o){for(var s=i.getText(),c=e(o,function(y,N){var Y=y.range.start.line-N.range.start.line;return Y===0?y.range.start.character-N.range.start.character:Y}),g=s.length,f=c.length-1;f>=0;f--){var p=c[f],b=i.offsetAt(p.range.start),d=i.offsetAt(p.range.end);if(d<=g)s=s.substring(0,b)+p.newText+s.substring(d,s.length);else throw new Error("Overlapping edit");g=b}return s}n.applyEdits=r;function e(i,o){if(i.length<=1)return i;var s=i.length/2|0,c=i.slice(0,s),g=i.slice(s);e(c,o),e(g,o);for(var f=0,p=0,b=0;f<c.length&&p<g.length;){var d=o(c[f],g[p]);d<=0?i[b++]=c[f++]:i[b++]=g[p++]}for(;f<c.length;)i[b++]=c[f++];for(;p<g.length;)i[b++]=g[p++];return i}})(De||(De={}));var Ne=function(){function n(t,a,r,e){this._uri=t,this._languageId=a,this._version=r,this._content=e,this._lineOffsets=void 0}return Object.defineProperty(n.prototype,"uri",{get:function(){return this._uri},enumerable:!1,configurable:!0}),Object.defineProperty(n.prototype,"languageId",{get:function(){return this._languageId},enumerable:!1,configurable:!0}),Object.defineProperty(n.prototype,"version",{get:function(){return this._version},enumerable:!1,configurable:!0}),n.prototype.getText=function(t){if(t){var a=this.offsetAt(t.start),r=this.offsetAt(t.end);return this._content.substring(a,r)}return this._content},n.prototype.update=function(t,a){this._content=t.text,this._version=a,this._lineOffsets=void 0},n.prototype.getLineOffsets=function(){if(this._lineOffsets===void 0){for(var t=[],a=this._content,r=!0,e=0;e<a.length;e++){r&&(t.push(e),r=!1);var i=a.charAt(e);r=i==="\r"||i===`
`,i==="\r"&&e+1<a.length&&a.charAt(e+1)===`
`&&e++}r&&a.length>0&&t.push(a.length),this._lineOffsets=t}return this._lineOffsets},n.prototype.positionAt=function(t){t=Math.max(Math.min(t,this._content.length),0);var a=this.getLineOffsets(),r=0,e=a.length;if(e===0)return k.create(0,t);for(;r<e;){var i=Math.floor((r+e)/2);a[i]>t?e=i:r=i+1}var o=r-1;return k.create(o,t-a[o])},n.prototype.offsetAt=function(t){var a=this.getLineOffsets();if(t.line>=a.length)return this._content.length;if(t.line<0)return 0;var r=a[t.line],e=t.line+1<a.length?a[t.line+1]:this._content.length;return Math.max(Math.min(r+t.character,e),r)},Object.defineProperty(n.prototype,"lineCount",{get:function(){return this.getLineOffsets().length},enumerable:!1,configurable:!0}),n}(),u;(function(n){var t=Object.prototype.toString;function a(d){return typeof d!="undefined"}n.defined=a;function r(d){return typeof d=="undefined"}n.undefined=r;function e(d){return d===!0||d===!1}n.boolean=e;function i(d){return t.call(d)==="[object String]"}n.string=i;function o(d){return t.call(d)==="[object Number]"}n.number=o;function s(d,y,N){return t.call(d)==="[object Number]"&&y<=d&&d<=N}n.numberRange=s;function c(d){return t.call(d)==="[object Number]"&&-2147483648<=d&&d<=2147483647}n.integer=c;function g(d){return t.call(d)==="[object Number]"&&0<=d&&d<=2147483647}n.uinteger=g;function f(d){return t.call(d)==="[object Function]"}n.func=f;function p(d){return d!==null&&typeof d=="object"}n.objectLiteral=p;function b(d,y){return Array.isArray(d)&&d.every(y)}n.typedArray=b})(u||(u={}));var We=function(){function n(t,a,r){var e=this;this._languageId=t,this._worker=a,this._disposables=[],this._listener=Object.create(null);var i=function(s){var c=s.getLanguageId();if(c===e._languageId){var g;e._listener[s.uri.toString()]=s.onDidChangeContent(function(){window.clearTimeout(g),g=window.setTimeout(function(){return e._doValidate(s.uri,c)},500)}),e._doValidate(s.uri,c)}},o=function(s){x.setModelMarkers(s,e._languageId,[]);var c=s.uri.toString(),g=e._listener[c];g&&(g.dispose(),delete e._listener[c])};this._disposables.push(x.onDidCreateModel(i)),this._disposables.push(x.onWillDisposeModel(o)),this._disposables.push(x.onDidChangeModelLanguage(function(s){o(s.model),i(s.model)})),r.onDidChange(function(s){x.getModels().forEach(function(c){c.getLanguageId()===e._languageId&&(o(c),i(c))})}),this._disposables.push({dispose:function(){for(var s in e._listener)e._listener[s].dispose()}}),x.getModels().forEach(i)}return n.prototype.dispose=function(){this._disposables.forEach(function(t){return t&&t.dispose()}),this._disposables=[]},n.prototype._doValidate=function(t,a){this._worker(t).then(function(r){return r.doValidation(t.toString())}).then(function(r){var e=r.map(function(o){return He(t,o)}),i=x.getModel(t);i&&i.getLanguageId()===a&&x.setModelMarkers(i,a,e)}).then(void 0,function(r){console.error(r)})},n}();function Ue(n){switch(n){case I.Error:return T.Error;case I.Warning:return T.Warning;case I.Information:return T.Info;case I.Hint:return T.Hint;default:return T.Info}}function He(n,t){var a=typeof t.code=="number"?String(t.code):t.code;return{severity:Ue(t.severity),startLineNumber:t.range.start.line+1,startColumn:t.range.start.character+1,endLineNumber:t.range.end.line+1,endColumn:t.range.end.character+1,message:t.message,code:a,source:t.source}}function C(n){if(!!n)return{character:n.column-1,line:n.lineNumber-1}}function Ve(n){if(!!n)return{start:{line:n.startLineNumber-1,character:n.startColumn-1},end:{line:n.endLineNumber-1,character:n.endColumn-1}}}function w(n){if(!!n)return new Z(n.start.line+1,n.start.character+1,n.end.line+1,n.end.character+1)}function Oe(n){return typeof n.insert!="undefined"&&typeof n.replace!="undefined"}function ze(n){var t=l.CompletionItemKind;switch(n){case h.Text:return t.Text;case h.Method:return t.Method;case h.Function:return t.Function;case h.Constructor:return t.Constructor;case h.Field:return t.Field;case h.Variable:return t.Variable;case h.Class:return t.Class;case h.Interface:return t.Interface;case h.Module:return t.Module;case h.Property:return t.Property;case h.Unit:return t.Unit;case h.Value:return t.Value;case h.Enum:return t.Enum;case h.Keyword:return t.Keyword;case h.Snippet:return t.Snippet;case h.Color:return t.Color;case h.File:return t.File;case h.Reference:return t.Reference}return t.Property}function J(n){if(!!n)return{range:w(n.range),text:n.newText}}function Xe(n){return n&&n.command==="editor.action.triggerSuggest"?{id:n.command,title:n.title,arguments:n.arguments}:void 0}var $e=function(){function n(t){this._worker=t}return Object.defineProperty(n.prototype,"triggerCharacters",{get:function(){return["/","-",":"]},enumerable:!1,configurable:!0}),n.prototype.provideCompletionItems=function(t,a,r,e){var i=t.uri;return this._worker(i).then(function(o){return o.doComplete(i.toString(),C(a))}).then(function(o){if(!!o){var s=t.getWordUntilPosition(a),c=new Z(a.lineNumber,s.startColumn,a.lineNumber,s.endColumn),g=o.items.map(function(f){var p={label:f.label,insertText:f.insertText||f.label,sortText:f.sortText,filterText:f.filterText,documentation:f.documentation,detail:f.detail,command:Xe(f.command),range:c,kind:ze(f.kind)};return f.textEdit&&(Oe(f.textEdit)?p.range={insert:w(f.textEdit.insert),replace:w(f.textEdit.replace)}:p.range=w(f.textEdit.range),p.insertText=f.textEdit.newText),f.additionalTextEdits&&(p.additionalTextEdits=f.additionalTextEdits.map(J)),f.insertTextFormat===G.Snippet&&(p.insertTextRules=l.CompletionItemInsertTextRule.InsertAsSnippet),p});return{isIncomplete:o.isIncomplete,suggestions:g}}})},n}();function Be(n){return n&&typeof n=="object"&&typeof n.kind=="string"}function Me(n){return typeof n=="string"?{value:n}:Be(n)?n.kind==="plaintext"?{value:n.value.replace(/[\\`*_{}[\]()#+\-.!]/g,"\\$&")}:{value:n.value}:{value:"```"+n.language+`
`+n.value+"\n```\n"}}function qe(n){if(!!n)return Array.isArray(n)?n.map(Me):[Me(n)]}var Qe=function(){function n(t){this._worker=t}return n.prototype.provideHover=function(t,a,r){var e=t.uri;return this._worker(e).then(function(i){return i.doHover(e.toString(),C(a))}).then(function(i){if(!!i)return{range:w(i.range),contents:qe(i.contents)}})},n}();function Ge(n){switch(n){case j.Read:return l.DocumentHighlightKind.Read;case j.Write:return l.DocumentHighlightKind.Write;case j.Text:return l.DocumentHighlightKind.Text}return l.DocumentHighlightKind.Text}var Je=function(){function n(t){this._worker=t}return n.prototype.provideDocumentHighlights=function(t,a,r){var e=t.uri;return this._worker(e).then(function(i){return i.findDocumentHighlights(e.toString(),C(a))}).then(function(i){if(!!i)return i.map(function(o){return{range:w(o.range),kind:Ge(o.kind)}})})},n}();function Pe(n){return{uri:K.parse(n.uri),range:w(n.range)}}var Ye=function(){function n(t){this._worker=t}return n.prototype.provideDefinition=function(t,a,r){var e=t.uri;return this._worker(e).then(function(i){return i.findDefinition(e.toString(),C(a))}).then(function(i){if(!!i)return[Pe(i)]})},n}(),Ze=function(){function n(t){this._worker=t}return n.prototype.provideReferences=function(t,a,r,e){var i=t.uri;return this._worker(i).then(function(o){return o.findReferences(i.toString(),C(a))}).then(function(o){if(!!o)return o.map(Pe)})},n}();function Ke(n){if(!(!n||!n.changes)){var t=[];for(var a in n.changes)for(var r=K.parse(a),e=0,i=n.changes[a];e<i.length;e++){var o=i[e];t.push({resource:r,edit:{range:w(o.range),text:o.newText}})}return{edits:t}}}var en=function(){function n(t){this._worker=t}return n.prototype.provideRenameEdits=function(t,a,r,e){var i=t.uri;return this._worker(i).then(function(o){return o.doRename(i.toString(),C(a),r)}).then(function(o){return Ke(o)})},n}();function nn(n){var t=l.SymbolKind;switch(n){case v.File:return t.Array;case v.Module:return t.Module;case v.Namespace:return t.Namespace;case v.Package:return t.Package;case v.Class:return t.Class;case v.Method:return t.Method;case v.Property:return t.Property;case v.Field:return t.Field;case v.Constructor:return t.Constructor;case v.Enum:return t.Enum;case v.Interface:return t.Interface;case v.Function:return t.Function;case v.Variable:return t.Variable;case v.Constant:return t.Constant;case v.String:return t.String;case v.Number:return t.Number;case v.Boolean:return t.Boolean;case v.Array:return t.Array}return t.Function}var tn=function(){function n(t){this._worker=t}return n.prototype.provideDocumentSymbols=function(t,a){var r=t.uri;return this._worker(r).then(function(e){return e.findDocumentSymbols(r.toString())}).then(function(e){if(!!e)return e.map(function(i){return{name:i.name,detail:"",containerName:i.containerName,kind:nn(i.kind),tags:[],range:w(i.location.range),selectionRange:w(i.location.range)}})})},n}(),rn=function(){function n(t){this._worker=t}return n.prototype.provideDocumentColors=function(t,a){var r=t.uri;return this._worker(r).then(function(e){return e.findDocumentColors(r.toString())}).then(function(e){if(!!e)return e.map(function(i){return{color:i.color,range:w(i.range)}})})},n.prototype.provideColorPresentations=function(t,a,r){var e=t.uri;return this._worker(e).then(function(i){return i.getColorPresentations(e.toString(),a.color,Ve(a.range))}).then(function(i){if(!!i)return i.map(function(o){var s={label:o.label};return o.textEdit&&(s.textEdit=J(o.textEdit)),o.additionalTextEdits&&(s.additionalTextEdits=o.additionalTextEdits.map(J)),s})})},n}(),an=function(){function n(t){this._worker=t}return n.prototype.provideFoldingRanges=function(t,a,r){var e=t.uri;return this._worker(e).then(function(i){return i.getFoldingRanges(e.toString(),a)}).then(function(i){if(!!i)return i.map(function(o){var s={start:o.startLine+1,end:o.endLine+1};return typeof o.kind!="undefined"&&(s.kind=on(o.kind)),s})})},n}();function on(n){switch(n){case D.Comment:return l.FoldingRangeKind.Comment;case D.Imports:return l.FoldingRangeKind.Imports;case D.Region:return l.FoldingRangeKind.Region}}var un=function(){function n(t){this._worker=t}return n.prototype.provideSelectionRanges=function(t,a,r){var e=t.uri;return this._worker(e).then(function(i){return i.getSelectionRanges(e.toString(),a.map(C))}).then(function(i){if(!!i)return i.map(function(o){for(var s=[];o;)s.push({range:w(o.range)}),o=o.parent;return s})})},n}();function cn(n){var t=[],a=[],r=new je(n);t.push(r);var e=function(){for(var o=[],s=0;s<arguments.length;s++)o[s]=arguments[s];return r.getLanguageServiceWorker.apply(r,o)};function i(){var o=n.languageId,s=n.modeConfiguration;Se(a),s.completionItems&&a.push(l.registerCompletionItemProvider(o,new $e(e))),s.hovers&&a.push(l.registerHoverProvider(o,new Qe(e))),s.documentHighlights&&a.push(l.registerDocumentHighlightProvider(o,new Je(e))),s.definitions&&a.push(l.registerDefinitionProvider(o,new Ye(e))),s.references&&a.push(l.registerReferenceProvider(o,new Ze(e))),s.documentSymbols&&a.push(l.registerDocumentSymbolProvider(o,new tn(e))),s.rename&&a.push(l.registerRenameProvider(o,new en(e))),s.colors&&a.push(l.registerColorProvider(o,new rn(e))),s.foldingRanges&&a.push(l.registerFoldingRangeProvider(o,new an(e))),s.diagnostics&&a.push(new We(o,e,n)),s.selectionRanges&&a.push(l.registerSelectionRangeProvider(o,new un(e)))}return i(),t.push(Le(a)),Le(t)}function Le(n){return{dispose:function(){return Se(n)}}}function Se(n){for(;n.length;)n.pop().dispose()}export{cn as setupMode};
