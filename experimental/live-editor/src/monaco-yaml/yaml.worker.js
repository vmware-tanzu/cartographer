/**
 * Copyright 2021 VMware
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// src/yaml.worker.ts
import { initialize } from "monaco-editor/esm/vs/editor/editor.worker.js";

// src/yamlWorker.ts
import { TextDocument as TextDocument2 } from "vscode-languageserver-textdocument";

// node_modules/vscode-json-languageservice/lib/esm/services/jsonSchemaService.js
import {
  parse
} from "jsonc-parser";

// node_modules/vscode-uri/lib/esm/index.js
var LIB;
LIB = (() => {
  "use strict";
  var t = { 470: (t2) => {
    function e2(t3) {
      if (typeof t3 != "string")
        throw new TypeError("Path must be a string. Received " + JSON.stringify(t3));
    }
    function r2(t3, e3) {
      for (var r3, n2 = "", o = 0, i = -1, a2 = 0, h = 0; h <= t3.length; ++h) {
        if (h < t3.length)
          r3 = t3.charCodeAt(h);
        else {
          if (r3 === 47)
            break;
          r3 = 47;
        }
        if (r3 === 47) {
          if (i === h - 1 || a2 === 1)
            ;
          else if (i !== h - 1 && a2 === 2) {
            if (n2.length < 2 || o !== 2 || n2.charCodeAt(n2.length - 1) !== 46 || n2.charCodeAt(n2.length - 2) !== 46) {
              if (n2.length > 2) {
                var s = n2.lastIndexOf("/");
                if (s !== n2.length - 1) {
                  s === -1 ? (n2 = "", o = 0) : o = (n2 = n2.slice(0, s)).length - 1 - n2.lastIndexOf("/"), i = h, a2 = 0;
                  continue;
                }
              } else if (n2.length === 2 || n2.length === 1) {
                n2 = "", o = 0, i = h, a2 = 0;
                continue;
              }
            }
            e3 && (n2.length > 0 ? n2 += "/.." : n2 = "..", o = 2);
          } else
            n2.length > 0 ? n2 += "/" + t3.slice(i + 1, h) : n2 = t3.slice(i + 1, h), o = h - i - 1;
          i = h, a2 = 0;
        } else
          r3 === 46 && a2 !== -1 ? ++a2 : a2 = -1;
      }
      return n2;
    }
    var n = { resolve: function() {
      for (var t3, n2 = "", o = false, i = arguments.length - 1; i >= -1 && !o; i--) {
        var a2;
        i >= 0 ? a2 = arguments[i] : (t3 === void 0 && (t3 = process.cwd()), a2 = t3), e2(a2), a2.length !== 0 && (n2 = a2 + "/" + n2, o = a2.charCodeAt(0) === 47);
      }
      return n2 = r2(n2, !o), o ? n2.length > 0 ? "/" + n2 : "/" : n2.length > 0 ? n2 : ".";
    }, normalize: function(t3) {
      if (e2(t3), t3.length === 0)
        return ".";
      var n2 = t3.charCodeAt(0) === 47, o = t3.charCodeAt(t3.length - 1) === 47;
      return (t3 = r2(t3, !n2)).length !== 0 || n2 || (t3 = "."), t3.length > 0 && o && (t3 += "/"), n2 ? "/" + t3 : t3;
    }, isAbsolute: function(t3) {
      return e2(t3), t3.length > 0 && t3.charCodeAt(0) === 47;
    }, join: function() {
      if (arguments.length === 0)
        return ".";
      for (var t3, r3 = 0; r3 < arguments.length; ++r3) {
        var o = arguments[r3];
        e2(o), o.length > 0 && (t3 === void 0 ? t3 = o : t3 += "/" + o);
      }
      return t3 === void 0 ? "." : n.normalize(t3);
    }, relative: function(t3, r3) {
      if (e2(t3), e2(r3), t3 === r3)
        return "";
      if ((t3 = n.resolve(t3)) === (r3 = n.resolve(r3)))
        return "";
      for (var o = 1; o < t3.length && t3.charCodeAt(o) === 47; ++o)
        ;
      for (var i = t3.length, a2 = i - o, h = 1; h < r3.length && r3.charCodeAt(h) === 47; ++h)
        ;
      for (var s = r3.length - h, f2 = a2 < s ? a2 : s, u = -1, c = 0; c <= f2; ++c) {
        if (c === f2) {
          if (s > f2) {
            if (r3.charCodeAt(h + c) === 47)
              return r3.slice(h + c + 1);
            if (c === 0)
              return r3.slice(h + c);
          } else
            a2 > f2 && (t3.charCodeAt(o + c) === 47 ? u = c : c === 0 && (u = 0));
          break;
        }
        var l = t3.charCodeAt(o + c);
        if (l !== r3.charCodeAt(h + c))
          break;
        l === 47 && (u = c);
      }
      var p = "";
      for (c = o + u + 1; c <= i; ++c)
        c !== i && t3.charCodeAt(c) !== 47 || (p.length === 0 ? p += ".." : p += "/..");
      return p.length > 0 ? p + r3.slice(h + u) : (h += u, r3.charCodeAt(h) === 47 && ++h, r3.slice(h));
    }, _makeLong: function(t3) {
      return t3;
    }, dirname: function(t3) {
      if (e2(t3), t3.length === 0)
        return ".";
      for (var r3 = t3.charCodeAt(0), n2 = r3 === 47, o = -1, i = true, a2 = t3.length - 1; a2 >= 1; --a2)
        if ((r3 = t3.charCodeAt(a2)) === 47) {
          if (!i) {
            o = a2;
            break;
          }
        } else
          i = false;
      return o === -1 ? n2 ? "/" : "." : n2 && o === 1 ? "//" : t3.slice(0, o);
    }, basename: function(t3, r3) {
      if (r3 !== void 0 && typeof r3 != "string")
        throw new TypeError('"ext" argument must be a string');
      e2(t3);
      var n2, o = 0, i = -1, a2 = true;
      if (r3 !== void 0 && r3.length > 0 && r3.length <= t3.length) {
        if (r3.length === t3.length && r3 === t3)
          return "";
        var h = r3.length - 1, s = -1;
        for (n2 = t3.length - 1; n2 >= 0; --n2) {
          var f2 = t3.charCodeAt(n2);
          if (f2 === 47) {
            if (!a2) {
              o = n2 + 1;
              break;
            }
          } else
            s === -1 && (a2 = false, s = n2 + 1), h >= 0 && (f2 === r3.charCodeAt(h) ? --h == -1 && (i = n2) : (h = -1, i = s));
        }
        return o === i ? i = s : i === -1 && (i = t3.length), t3.slice(o, i);
      }
      for (n2 = t3.length - 1; n2 >= 0; --n2)
        if (t3.charCodeAt(n2) === 47) {
          if (!a2) {
            o = n2 + 1;
            break;
          }
        } else
          i === -1 && (a2 = false, i = n2 + 1);
      return i === -1 ? "" : t3.slice(o, i);
    }, extname: function(t3) {
      e2(t3);
      for (var r3 = -1, n2 = 0, o = -1, i = true, a2 = 0, h = t3.length - 1; h >= 0; --h) {
        var s = t3.charCodeAt(h);
        if (s !== 47)
          o === -1 && (i = false, o = h + 1), s === 46 ? r3 === -1 ? r3 = h : a2 !== 1 && (a2 = 1) : r3 !== -1 && (a2 = -1);
        else if (!i) {
          n2 = h + 1;
          break;
        }
      }
      return r3 === -1 || o === -1 || a2 === 0 || a2 === 1 && r3 === o - 1 && r3 === n2 + 1 ? "" : t3.slice(r3, o);
    }, format: function(t3) {
      if (t3 === null || typeof t3 != "object")
        throw new TypeError('The "pathObject" argument must be of type Object. Received type ' + typeof t3);
      return function(t4, e3) {
        var r3 = e3.dir || e3.root, n2 = e3.base || (e3.name || "") + (e3.ext || "");
        return r3 ? r3 === e3.root ? r3 + n2 : r3 + "/" + n2 : n2;
      }(0, t3);
    }, parse: function(t3) {
      e2(t3);
      var r3 = { root: "", dir: "", base: "", ext: "", name: "" };
      if (t3.length === 0)
        return r3;
      var n2, o = t3.charCodeAt(0), i = o === 47;
      i ? (r3.root = "/", n2 = 1) : n2 = 0;
      for (var a2 = -1, h = 0, s = -1, f2 = true, u = t3.length - 1, c = 0; u >= n2; --u)
        if ((o = t3.charCodeAt(u)) !== 47)
          s === -1 && (f2 = false, s = u + 1), o === 46 ? a2 === -1 ? a2 = u : c !== 1 && (c = 1) : a2 !== -1 && (c = -1);
        else if (!f2) {
          h = u + 1;
          break;
        }
      return a2 === -1 || s === -1 || c === 0 || c === 1 && a2 === s - 1 && a2 === h + 1 ? s !== -1 && (r3.base = r3.name = h === 0 && i ? t3.slice(1, s) : t3.slice(h, s)) : (h === 0 && i ? (r3.name = t3.slice(1, a2), r3.base = t3.slice(1, s)) : (r3.name = t3.slice(h, a2), r3.base = t3.slice(h, s)), r3.ext = t3.slice(a2, s)), h > 0 ? r3.dir = t3.slice(0, h - 1) : i && (r3.dir = "/"), r3;
    }, sep: "/", delimiter: ":", win32: null, posix: null };
    n.posix = n, t2.exports = n;
  }, 447: (t2, e2, r2) => {
    var n;
    if (r2.r(e2), r2.d(e2, { URI: () => g, Utils: () => O }), typeof process == "object")
      n = process.platform === "win32";
    else if (typeof navigator == "object") {
      var o = navigator.userAgent;
      n = o.indexOf("Windows") >= 0;
    }
    var i, a2, h = (i = function(t3, e3) {
      return (i = Object.setPrototypeOf || { __proto__: [] } instanceof Array && function(t4, e4) {
        t4.__proto__ = e4;
      } || function(t4, e4) {
        for (var r3 in e4)
          Object.prototype.hasOwnProperty.call(e4, r3) && (t4[r3] = e4[r3]);
      })(t3, e3);
    }, function(t3, e3) {
      function r3() {
        this.constructor = t3;
      }
      i(t3, e3), t3.prototype = e3 === null ? Object.create(e3) : (r3.prototype = e3.prototype, new r3());
    }), s = /^\w[\w\d+.-]*$/, f2 = /^\//, u = /^\/\//, c = "", l = "/", p = /^(([^:/?#]+?):)?(\/\/([^/?#]*))?([^?#]*)(\?([^#]*))?(#(.*))?/, g = function() {
      function t3(t4, e3, r3, n2, o2, i2) {
        i2 === void 0 && (i2 = false), typeof t4 == "object" ? (this.scheme = t4.scheme || c, this.authority = t4.authority || c, this.path = t4.path || c, this.query = t4.query || c, this.fragment = t4.fragment || c) : (this.scheme = function(t5, e4) {
          return t5 || e4 ? t5 : "file";
        }(t4, i2), this.authority = e3 || c, this.path = function(t5, e4) {
          switch (t5) {
            case "https":
            case "http":
            case "file":
              e4 ? e4[0] !== l && (e4 = l + e4) : e4 = l;
          }
          return e4;
        }(this.scheme, r3 || c), this.query = n2 || c, this.fragment = o2 || c, function(t5, e4) {
          if (!t5.scheme && e4)
            throw new Error('[UriError]: Scheme is missing: {scheme: "", authority: "' + t5.authority + '", path: "' + t5.path + '", query: "' + t5.query + '", fragment: "' + t5.fragment + '"}');
          if (t5.scheme && !s.test(t5.scheme))
            throw new Error("[UriError]: Scheme contains illegal characters.");
          if (t5.path) {
            if (t5.authority) {
              if (!f2.test(t5.path))
                throw new Error('[UriError]: If a URI contains an authority component, then the path component must either be empty or begin with a slash ("/") character');
            } else if (u.test(t5.path))
              throw new Error('[UriError]: If a URI does not contain an authority component, then the path cannot begin with two slash characters ("//")');
          }
        }(this, i2));
      }
      return t3.isUri = function(e3) {
        return e3 instanceof t3 || !!e3 && typeof e3.authority == "string" && typeof e3.fragment == "string" && typeof e3.path == "string" && typeof e3.query == "string" && typeof e3.scheme == "string" && typeof e3.fsPath == "function" && typeof e3.with == "function" && typeof e3.toString == "function";
      }, Object.defineProperty(t3.prototype, "fsPath", { get: function() {
        return C(this, false);
      }, enumerable: false, configurable: true }), t3.prototype.with = function(t4) {
        if (!t4)
          return this;
        var e3 = t4.scheme, r3 = t4.authority, n2 = t4.path, o2 = t4.query, i2 = t4.fragment;
        return e3 === void 0 ? e3 = this.scheme : e3 === null && (e3 = c), r3 === void 0 ? r3 = this.authority : r3 === null && (r3 = c), n2 === void 0 ? n2 = this.path : n2 === null && (n2 = c), o2 === void 0 ? o2 = this.query : o2 === null && (o2 = c), i2 === void 0 ? i2 = this.fragment : i2 === null && (i2 = c), e3 === this.scheme && r3 === this.authority && n2 === this.path && o2 === this.query && i2 === this.fragment ? this : new v(e3, r3, n2, o2, i2);
      }, t3.parse = function(t4, e3) {
        e3 === void 0 && (e3 = false);
        var r3 = p.exec(t4);
        return r3 ? new v(r3[2] || c, x(r3[4] || c), x(r3[5] || c), x(r3[7] || c), x(r3[9] || c), e3) : new v(c, c, c, c, c);
      }, t3.file = function(t4) {
        var e3 = c;
        if (n && (t4 = t4.replace(/\\/g, l)), t4[0] === l && t4[1] === l) {
          var r3 = t4.indexOf(l, 2);
          r3 === -1 ? (e3 = t4.substring(2), t4 = l) : (e3 = t4.substring(2, r3), t4 = t4.substring(r3) || l);
        }
        return new v("file", e3, t4, c, c);
      }, t3.from = function(t4) {
        return new v(t4.scheme, t4.authority, t4.path, t4.query, t4.fragment);
      }, t3.prototype.toString = function(t4) {
        return t4 === void 0 && (t4 = false), A2(this, t4);
      }, t3.prototype.toJSON = function() {
        return this;
      }, t3.revive = function(e3) {
        if (e3) {
          if (e3 instanceof t3)
            return e3;
          var r3 = new v(e3);
          return r3._formatted = e3.external, r3._fsPath = e3._sep === d ? e3.fsPath : null, r3;
        }
        return e3;
      }, t3;
    }(), d = n ? 1 : void 0, v = function(t3) {
      function e3() {
        var e4 = t3 !== null && t3.apply(this, arguments) || this;
        return e4._formatted = null, e4._fsPath = null, e4;
      }
      return h(e3, t3), Object.defineProperty(e3.prototype, "fsPath", { get: function() {
        return this._fsPath || (this._fsPath = C(this, false)), this._fsPath;
      }, enumerable: false, configurable: true }), e3.prototype.toString = function(t4) {
        return t4 === void 0 && (t4 = false), t4 ? A2(this, true) : (this._formatted || (this._formatted = A2(this, false)), this._formatted);
      }, e3.prototype.toJSON = function() {
        var t4 = { $mid: 1 };
        return this._fsPath && (t4.fsPath = this._fsPath, t4._sep = d), this._formatted && (t4.external = this._formatted), this.path && (t4.path = this.path), this.scheme && (t4.scheme = this.scheme), this.authority && (t4.authority = this.authority), this.query && (t4.query = this.query), this.fragment && (t4.fragment = this.fragment), t4;
      }, e3;
    }(g), m = ((a2 = {})[58] = "%3A", a2[47] = "%2F", a2[63] = "%3F", a2[35] = "%23", a2[91] = "%5B", a2[93] = "%5D", a2[64] = "%40", a2[33] = "%21", a2[36] = "%24", a2[38] = "%26", a2[39] = "%27", a2[40] = "%28", a2[41] = "%29", a2[42] = "%2A", a2[43] = "%2B", a2[44] = "%2C", a2[59] = "%3B", a2[61] = "%3D", a2[32] = "%20", a2);
    function y(t3, e3) {
      for (var r3 = void 0, n2 = -1, o2 = 0; o2 < t3.length; o2++) {
        var i2 = t3.charCodeAt(o2);
        if (i2 >= 97 && i2 <= 122 || i2 >= 65 && i2 <= 90 || i2 >= 48 && i2 <= 57 || i2 === 45 || i2 === 46 || i2 === 95 || i2 === 126 || e3 && i2 === 47)
          n2 !== -1 && (r3 += encodeURIComponent(t3.substring(n2, o2)), n2 = -1), r3 !== void 0 && (r3 += t3.charAt(o2));
        else {
          r3 === void 0 && (r3 = t3.substr(0, o2));
          var a3 = m[i2];
          a3 !== void 0 ? (n2 !== -1 && (r3 += encodeURIComponent(t3.substring(n2, o2)), n2 = -1), r3 += a3) : n2 === -1 && (n2 = o2);
        }
      }
      return n2 !== -1 && (r3 += encodeURIComponent(t3.substring(n2))), r3 !== void 0 ? r3 : t3;
    }
    function b(t3) {
      for (var e3 = void 0, r3 = 0; r3 < t3.length; r3++) {
        var n2 = t3.charCodeAt(r3);
        n2 === 35 || n2 === 63 ? (e3 === void 0 && (e3 = t3.substr(0, r3)), e3 += m[n2]) : e3 !== void 0 && (e3 += t3[r3]);
      }
      return e3 !== void 0 ? e3 : t3;
    }
    function C(t3, e3) {
      var r3;
      return r3 = t3.authority && t3.path.length > 1 && t3.scheme === "file" ? "//" + t3.authority + t3.path : t3.path.charCodeAt(0) === 47 && (t3.path.charCodeAt(1) >= 65 && t3.path.charCodeAt(1) <= 90 || t3.path.charCodeAt(1) >= 97 && t3.path.charCodeAt(1) <= 122) && t3.path.charCodeAt(2) === 58 ? e3 ? t3.path.substr(1) : t3.path[1].toLowerCase() + t3.path.substr(2) : t3.path, n && (r3 = r3.replace(/\//g, "\\")), r3;
    }
    function A2(t3, e3) {
      var r3 = e3 ? b : y, n2 = "", o2 = t3.scheme, i2 = t3.authority, a3 = t3.path, h2 = t3.query, s2 = t3.fragment;
      if (o2 && (n2 += o2, n2 += ":"), (i2 || o2 === "file") && (n2 += l, n2 += l), i2) {
        var f3 = i2.indexOf("@");
        if (f3 !== -1) {
          var u2 = i2.substr(0, f3);
          i2 = i2.substr(f3 + 1), (f3 = u2.indexOf(":")) === -1 ? n2 += r3(u2, false) : (n2 += r3(u2.substr(0, f3), false), n2 += ":", n2 += r3(u2.substr(f3 + 1), false)), n2 += "@";
        }
        (f3 = (i2 = i2.toLowerCase()).indexOf(":")) === -1 ? n2 += r3(i2, false) : (n2 += r3(i2.substr(0, f3), false), n2 += i2.substr(f3));
      }
      if (a3) {
        if (a3.length >= 3 && a3.charCodeAt(0) === 47 && a3.charCodeAt(2) === 58)
          (c2 = a3.charCodeAt(1)) >= 65 && c2 <= 90 && (a3 = "/" + String.fromCharCode(c2 + 32) + ":" + a3.substr(3));
        else if (a3.length >= 2 && a3.charCodeAt(1) === 58) {
          var c2;
          (c2 = a3.charCodeAt(0)) >= 65 && c2 <= 90 && (a3 = String.fromCharCode(c2 + 32) + ":" + a3.substr(2));
        }
        n2 += r3(a3, true);
      }
      return h2 && (n2 += "?", n2 += r3(h2, false)), s2 && (n2 += "#", n2 += e3 ? s2 : y(s2, false)), n2;
    }
    function w(t3) {
      try {
        return decodeURIComponent(t3);
      } catch (e3) {
        return t3.length > 3 ? t3.substr(0, 3) + w(t3.substr(3)) : t3;
      }
    }
    var _ = /(%[0-9A-Za-z][0-9A-Za-z])+/g;
    function x(t3) {
      return t3.match(_) ? t3.replace(_, function(t4) {
        return w(t4);
      }) : t3;
    }
    var O, P = r2(470), j = function() {
      for (var t3 = 0, e3 = 0, r3 = arguments.length; e3 < r3; e3++)
        t3 += arguments[e3].length;
      var n2 = Array(t3), o2 = 0;
      for (e3 = 0; e3 < r3; e3++)
        for (var i2 = arguments[e3], a3 = 0, h2 = i2.length; a3 < h2; a3++, o2++)
          n2[o2] = i2[a3];
      return n2;
    }, U = P.posix || P;
    !function(t3) {
      t3.joinPath = function(t4) {
        for (var e3 = [], r3 = 1; r3 < arguments.length; r3++)
          e3[r3 - 1] = arguments[r3];
        return t4.with({ path: U.join.apply(U, j([t4.path], e3)) });
      }, t3.resolvePath = function(t4) {
        for (var e3 = [], r3 = 1; r3 < arguments.length; r3++)
          e3[r3 - 1] = arguments[r3];
        var n2 = t4.path || "/";
        return t4.with({ path: U.resolve.apply(U, j([n2], e3)) });
      }, t3.dirname = function(t4) {
        var e3 = U.dirname(t4.path);
        return e3.length === 1 && e3.charCodeAt(0) === 46 ? t4 : t4.with({ path: e3 });
      }, t3.basename = function(t4) {
        return U.basename(t4.path);
      }, t3.extname = function(t4) {
        return U.extname(t4.path);
      };
    }(O || (O = {}));
  } }, e = {};
  function r(n) {
    if (e[n])
      return e[n].exports;
    var o = e[n] = { exports: {} };
    return t[n](o, o.exports, r), o.exports;
  }
  return r.d = (t2, e2) => {
    for (var n in e2)
      r.o(e2, n) && !r.o(t2, n) && Object.defineProperty(t2, n, { enumerable: true, get: e2[n] });
  }, r.o = (t2, e2) => Object.prototype.hasOwnProperty.call(t2, e2), r.r = (t2) => {
    typeof Symbol != "undefined" && Symbol.toStringTag && Object.defineProperty(t2, Symbol.toStringTag, { value: "Module" }), Object.defineProperty(t2, "__esModule", { value: true });
  }, r(447);
})();
var { URI, Utils } = LIB;

// node_modules/vscode-json-languageservice/lib/esm/utils/strings.js
function startsWith(haystack, needle) {
  if (haystack.length < needle.length) {
    return false;
  }
  for (var i = 0; i < needle.length; i++) {
    if (haystack[i] !== needle[i]) {
      return false;
    }
  }
  return true;
}
function endsWith(haystack, needle) {
  var diff = haystack.length - needle.length;
  if (diff > 0) {
    return haystack.lastIndexOf(needle) === diff;
  } else if (diff === 0) {
    return haystack === needle;
  } else {
    return false;
  }
}
function extendedRegExp(pattern) {
  var flags = "";
  if (startsWith(pattern, "(?i)")) {
    pattern = pattern.substring(4);
    flags = "i";
  }
  try {
    return new RegExp(pattern, flags + "u");
  } catch (e) {
    try {
      return new RegExp(pattern, flags);
    } catch (e2) {
      return void 0;
    }
  }
}

// node_modules/vscode-json-languageservice/lib/esm/parser/jsonParser.js
import {
  createScanner,
  findNodeAtOffset,
  getNodePath,
  getNodeValue
} from "jsonc-parser";

// node_modules/vscode-json-languageservice/lib/esm/utils/objects.js
function equals(one, other) {
  if (one === other) {
    return true;
  }
  if (one === null || one === void 0 || other === null || other === void 0) {
    return false;
  }
  if (typeof one !== typeof other) {
    return false;
  }
  if (typeof one !== "object") {
    return false;
  }
  if (Array.isArray(one) !== Array.isArray(other)) {
    return false;
  }
  var i, key;
  if (Array.isArray(one)) {
    if (one.length !== other.length) {
      return false;
    }
    for (i = 0; i < one.length; i++) {
      if (!equals(one[i], other[i])) {
        return false;
      }
    }
  } else {
    var oneKeys = [];
    for (key in one) {
      oneKeys.push(key);
    }
    oneKeys.sort();
    var otherKeys = [];
    for (key in other) {
      otherKeys.push(key);
    }
    otherKeys.sort();
    if (!equals(oneKeys, otherKeys)) {
      return false;
    }
    for (i = 0; i < oneKeys.length; i++) {
      if (!equals(one[oneKeys[i]], other[oneKeys[i]])) {
        return false;
      }
    }
  }
  return true;
}
function isNumber(val) {
  return typeof val === "number";
}
function isDefined(val) {
  return typeof val !== "undefined";
}
function isBoolean(val) {
  return typeof val === "boolean";
}
function isString(val) {
  return typeof val === "string";
}

// node_modules/vscode-json-languageservice/lib/esm/jsonLanguageTypes.js
import { Range, Position, MarkupContent, MarkupKind, Color, ColorInformation, ColorPresentation, FoldingRange, FoldingRangeKind, SelectionRange, Diagnostic, DiagnosticSeverity, CompletionItem, CompletionItemKind, CompletionList, CompletionItemTag, InsertTextFormat, SymbolInformation, SymbolKind, DocumentSymbol, Location, Hover, MarkedString, CodeActionContext, Command, CodeAction, DocumentHighlight, DocumentLink, WorkspaceEdit, TextEdit, CodeActionKind, TextDocumentEdit, VersionedTextDocumentIdentifier, DocumentHighlightKind } from "vscode-languageserver-types";
import { TextDocument } from "vscode-languageserver-textdocument";
var ErrorCode;
(function(ErrorCode2) {
  ErrorCode2[ErrorCode2["Undefined"] = 0] = "Undefined";
  ErrorCode2[ErrorCode2["EnumValueMismatch"] = 1] = "EnumValueMismatch";
  ErrorCode2[ErrorCode2["Deprecated"] = 2] = "Deprecated";
  ErrorCode2[ErrorCode2["UnexpectedEndOfComment"] = 257] = "UnexpectedEndOfComment";
  ErrorCode2[ErrorCode2["UnexpectedEndOfString"] = 258] = "UnexpectedEndOfString";
  ErrorCode2[ErrorCode2["UnexpectedEndOfNumber"] = 259] = "UnexpectedEndOfNumber";
  ErrorCode2[ErrorCode2["InvalidUnicode"] = 260] = "InvalidUnicode";
  ErrorCode2[ErrorCode2["InvalidEscapeCharacter"] = 261] = "InvalidEscapeCharacter";
  ErrorCode2[ErrorCode2["InvalidCharacter"] = 262] = "InvalidCharacter";
  ErrorCode2[ErrorCode2["PropertyExpected"] = 513] = "PropertyExpected";
  ErrorCode2[ErrorCode2["CommaExpected"] = 514] = "CommaExpected";
  ErrorCode2[ErrorCode2["ColonExpected"] = 515] = "ColonExpected";
  ErrorCode2[ErrorCode2["ValueExpected"] = 516] = "ValueExpected";
  ErrorCode2[ErrorCode2["CommaOrCloseBacketExpected"] = 517] = "CommaOrCloseBacketExpected";
  ErrorCode2[ErrorCode2["CommaOrCloseBraceExpected"] = 518] = "CommaOrCloseBraceExpected";
  ErrorCode2[ErrorCode2["TrailingComma"] = 519] = "TrailingComma";
  ErrorCode2[ErrorCode2["DuplicateKey"] = 520] = "DuplicateKey";
  ErrorCode2[ErrorCode2["CommentNotPermitted"] = 521] = "CommentNotPermitted";
  ErrorCode2[ErrorCode2["SchemaResolveError"] = 768] = "SchemaResolveError";
})(ErrorCode || (ErrorCode = {}));
var ClientCapabilities;
(function(ClientCapabilities2) {
  ClientCapabilities2.LATEST = {
    textDocument: {
      completion: {
        completionItem: {
          documentationFormat: [MarkupKind.Markdown, MarkupKind.PlainText],
          commitCharactersSupport: true
        }
      }
    }
  };
})(ClientCapabilities || (ClientCapabilities = {}));

// src/fillers/vscode-nls.ts
function format(message, args) {
  return args.length === 0 ? message : message.replace(/{(\d+)}/g, (match, rest) => {
    const [index] = rest;
    return typeof args[index] === "undefined" ? match : args[index];
  });
}
function localize(key, message, ...args) {
  return format(message, args);
}
function loadMessageBundle() {
  return localize;
}

// node_modules/vscode-json-languageservice/lib/esm/parser/jsonParser.js
var __extends = function() {
  var extendStatics = function(d, b) {
    extendStatics = Object.setPrototypeOf || { __proto__: [] } instanceof Array && function(d2, b2) {
      d2.__proto__ = b2;
    } || function(d2, b2) {
      for (var p in b2)
        if (Object.prototype.hasOwnProperty.call(b2, p))
          d2[p] = b2[p];
    };
    return extendStatics(d, b);
  };
  return function(d, b) {
    if (typeof b !== "function" && b !== null)
      throw new TypeError("Class extends value " + String(b) + " is not a constructor or null");
    extendStatics(d, b);
    function __() {
      this.constructor = d;
    }
    d.prototype = b === null ? Object.create(b) : (__.prototype = b.prototype, new __());
  };
}();
var localize2 = loadMessageBundle();
var formats = {
  "color-hex": { errorMessage: localize2("colorHexFormatWarning", "Invalid color format. Use #RGB, #RGBA, #RRGGBB or #RRGGBBAA."), pattern: /^#([0-9A-Fa-f]{3,4}|([0-9A-Fa-f]{2}){3,4})$/ },
  "date-time": { errorMessage: localize2("dateTimeFormatWarning", "String is not a RFC3339 date-time."), pattern: /^(\d{4})-(0[1-9]|1[0-2])-(0[1-9]|[12][0-9]|3[01])T([01][0-9]|2[0-3]):([0-5][0-9]):([0-5][0-9]|60)(\.[0-9]+)?(Z|(\+|-)([01][0-9]|2[0-3]):([0-5][0-9]))$/i },
  "date": { errorMessage: localize2("dateFormatWarning", "String is not a RFC3339 date."), pattern: /^(\d{4})-(0[1-9]|1[0-2])-(0[1-9]|[12][0-9]|3[01])$/i },
  "time": { errorMessage: localize2("timeFormatWarning", "String is not a RFC3339 time."), pattern: /^([01][0-9]|2[0-3]):([0-5][0-9]):([0-5][0-9]|60)(\.[0-9]+)?(Z|(\+|-)([01][0-9]|2[0-3]):([0-5][0-9]))$/i },
  "email": { errorMessage: localize2("emailFormatWarning", "String is not an e-mail address."), pattern: /^(([^<>()\[\]\\.,;:\s@"]+(\.[^<>()\[\]\\.,;:\s@"]+)*)|(".+"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$/ }
};
var ASTNodeImpl = function() {
  function ASTNodeImpl3(parent, offset, length) {
    if (length === void 0) {
      length = 0;
    }
    this.offset = offset;
    this.length = length;
    this.parent = parent;
  }
  Object.defineProperty(ASTNodeImpl3.prototype, "children", {
    get: function() {
      return [];
    },
    enumerable: false,
    configurable: true
  });
  ASTNodeImpl3.prototype.toString = function() {
    return "type: " + this.type + " (" + this.offset + "/" + this.length + ")" + (this.parent ? " parent: {" + this.parent.toString() + "}" : "");
  };
  return ASTNodeImpl3;
}();
var NullASTNodeImpl = function(_super) {
  __extends(NullASTNodeImpl3, _super);
  function NullASTNodeImpl3(parent, offset) {
    var _this = _super.call(this, parent, offset) || this;
    _this.type = "null";
    _this.value = null;
    return _this;
  }
  return NullASTNodeImpl3;
}(ASTNodeImpl);
var BooleanASTNodeImpl = function(_super) {
  __extends(BooleanASTNodeImpl3, _super);
  function BooleanASTNodeImpl3(parent, boolValue, offset) {
    var _this = _super.call(this, parent, offset) || this;
    _this.type = "boolean";
    _this.value = boolValue;
    return _this;
  }
  return BooleanASTNodeImpl3;
}(ASTNodeImpl);
var ArrayASTNodeImpl = function(_super) {
  __extends(ArrayASTNodeImpl3, _super);
  function ArrayASTNodeImpl3(parent, offset) {
    var _this = _super.call(this, parent, offset) || this;
    _this.type = "array";
    _this.items = [];
    return _this;
  }
  Object.defineProperty(ArrayASTNodeImpl3.prototype, "children", {
    get: function() {
      return this.items;
    },
    enumerable: false,
    configurable: true
  });
  return ArrayASTNodeImpl3;
}(ASTNodeImpl);
var NumberASTNodeImpl = function(_super) {
  __extends(NumberASTNodeImpl3, _super);
  function NumberASTNodeImpl3(parent, offset) {
    var _this = _super.call(this, parent, offset) || this;
    _this.type = "number";
    _this.isInteger = true;
    _this.value = Number.NaN;
    return _this;
  }
  return NumberASTNodeImpl3;
}(ASTNodeImpl);
var StringASTNodeImpl = function(_super) {
  __extends(StringASTNodeImpl3, _super);
  function StringASTNodeImpl3(parent, offset, length) {
    var _this = _super.call(this, parent, offset, length) || this;
    _this.type = "string";
    _this.value = "";
    return _this;
  }
  return StringASTNodeImpl3;
}(ASTNodeImpl);
var PropertyASTNodeImpl = function(_super) {
  __extends(PropertyASTNodeImpl3, _super);
  function PropertyASTNodeImpl3(parent, offset, keyNode) {
    var _this = _super.call(this, parent, offset) || this;
    _this.type = "property";
    _this.colonOffset = -1;
    _this.keyNode = keyNode;
    return _this;
  }
  Object.defineProperty(PropertyASTNodeImpl3.prototype, "children", {
    get: function() {
      return this.valueNode ? [this.keyNode, this.valueNode] : [this.keyNode];
    },
    enumerable: false,
    configurable: true
  });
  return PropertyASTNodeImpl3;
}(ASTNodeImpl);
var ObjectASTNodeImpl = function(_super) {
  __extends(ObjectASTNodeImpl3, _super);
  function ObjectASTNodeImpl3(parent, offset) {
    var _this = _super.call(this, parent, offset) || this;
    _this.type = "object";
    _this.properties = [];
    return _this;
  }
  Object.defineProperty(ObjectASTNodeImpl3.prototype, "children", {
    get: function() {
      return this.properties;
    },
    enumerable: false,
    configurable: true
  });
  return ObjectASTNodeImpl3;
}(ASTNodeImpl);
function asSchema(schema) {
  if (isBoolean(schema)) {
    return schema ? {} : { "not": {} };
  }
  return schema;
}
var EnumMatch;
(function(EnumMatch3) {
  EnumMatch3[EnumMatch3["Key"] = 0] = "Key";
  EnumMatch3[EnumMatch3["Enum"] = 1] = "Enum";
})(EnumMatch || (EnumMatch = {}));
var SchemaCollector = function() {
  function SchemaCollector3(focusOffset, exclude) {
    if (focusOffset === void 0) {
      focusOffset = -1;
    }
    this.focusOffset = focusOffset;
    this.exclude = exclude;
    this.schemas = [];
  }
  SchemaCollector3.prototype.add = function(schema) {
    this.schemas.push(schema);
  };
  SchemaCollector3.prototype.merge = function(other) {
    Array.prototype.push.apply(this.schemas, other.schemas);
  };
  SchemaCollector3.prototype.include = function(node) {
    return (this.focusOffset === -1 || contains(node, this.focusOffset)) && node !== this.exclude;
  };
  SchemaCollector3.prototype.newSub = function() {
    return new SchemaCollector3(-1, this.exclude);
  };
  return SchemaCollector3;
}();
var NoOpSchemaCollector = function() {
  function NoOpSchemaCollector3() {
  }
  Object.defineProperty(NoOpSchemaCollector3.prototype, "schemas", {
    get: function() {
      return [];
    },
    enumerable: false,
    configurable: true
  });
  NoOpSchemaCollector3.prototype.add = function(schema) {
  };
  NoOpSchemaCollector3.prototype.merge = function(other) {
  };
  NoOpSchemaCollector3.prototype.include = function(node) {
    return true;
  };
  NoOpSchemaCollector3.prototype.newSub = function() {
    return this;
  };
  NoOpSchemaCollector3.instance = new NoOpSchemaCollector3();
  return NoOpSchemaCollector3;
}();
var ValidationResult = function() {
  function ValidationResult3() {
    this.problems = [];
    this.propertiesMatches = 0;
    this.propertiesValueMatches = 0;
    this.primaryValueMatches = 0;
    this.enumValueMatch = false;
    this.enumValues = void 0;
  }
  ValidationResult3.prototype.hasProblems = function() {
    return !!this.problems.length;
  };
  ValidationResult3.prototype.mergeAll = function(validationResults) {
    for (var _i = 0, validationResults_1 = validationResults; _i < validationResults_1.length; _i++) {
      var validationResult = validationResults_1[_i];
      this.merge(validationResult);
    }
  };
  ValidationResult3.prototype.merge = function(validationResult) {
    this.problems = this.problems.concat(validationResult.problems);
  };
  ValidationResult3.prototype.mergeEnumValues = function(validationResult) {
    if (!this.enumValueMatch && !validationResult.enumValueMatch && this.enumValues && validationResult.enumValues) {
      this.enumValues = this.enumValues.concat(validationResult.enumValues);
      for (var _i = 0, _a = this.problems; _i < _a.length; _i++) {
        var error = _a[_i];
        if (error.code === ErrorCode.EnumValueMismatch) {
          error.message = localize2("enumWarning", "Value is not accepted. Valid values: {0}.", this.enumValues.map(function(v) {
            return JSON.stringify(v);
          }).join(", "));
        }
      }
    }
  };
  ValidationResult3.prototype.mergePropertyMatch = function(propertyValidationResult) {
    this.merge(propertyValidationResult);
    this.propertiesMatches++;
    if (propertyValidationResult.enumValueMatch || !propertyValidationResult.hasProblems() && propertyValidationResult.propertiesMatches) {
      this.propertiesValueMatches++;
    }
    if (propertyValidationResult.enumValueMatch && propertyValidationResult.enumValues && propertyValidationResult.enumValues.length === 1) {
      this.primaryValueMatches++;
    }
  };
  ValidationResult3.prototype.compare = function(other) {
    var hasProblems = this.hasProblems();
    if (hasProblems !== other.hasProblems()) {
      return hasProblems ? -1 : 1;
    }
    if (this.enumValueMatch !== other.enumValueMatch) {
      return other.enumValueMatch ? -1 : 1;
    }
    if (this.primaryValueMatches !== other.primaryValueMatches) {
      return this.primaryValueMatches - other.primaryValueMatches;
    }
    if (this.propertiesValueMatches !== other.propertiesValueMatches) {
      return this.propertiesValueMatches - other.propertiesValueMatches;
    }
    return this.propertiesMatches - other.propertiesMatches;
  };
  return ValidationResult3;
}();
function getNodeValue2(node) {
  return getNodeValue(node);
}
function getNodePath2(node) {
  return getNodePath(node);
}
function contains(node, offset, includeRightBound) {
  if (includeRightBound === void 0) {
    includeRightBound = false;
  }
  return offset >= node.offset && offset < node.offset + node.length || includeRightBound && offset === node.offset + node.length;
}
var JSONDocument = function() {
  function JSONDocument3(root, syntaxErrors, comments) {
    if (syntaxErrors === void 0) {
      syntaxErrors = [];
    }
    if (comments === void 0) {
      comments = [];
    }
    this.root = root;
    this.syntaxErrors = syntaxErrors;
    this.comments = comments;
  }
  JSONDocument3.prototype.getNodeFromOffset = function(offset, includeRightBound) {
    if (includeRightBound === void 0) {
      includeRightBound = false;
    }
    if (this.root) {
      return findNodeAtOffset(this.root, offset, includeRightBound);
    }
    return void 0;
  };
  JSONDocument3.prototype.visit = function(visitor) {
    if (this.root) {
      var doVisit_1 = function(node) {
        var ctn = visitor(node);
        var children = node.children;
        if (Array.isArray(children)) {
          for (var i = 0; i < children.length && ctn; i++) {
            ctn = doVisit_1(children[i]);
          }
        }
        return ctn;
      };
      doVisit_1(this.root);
    }
  };
  JSONDocument3.prototype.validate = function(textDocument, schema, severity) {
    if (severity === void 0) {
      severity = DiagnosticSeverity.Warning;
    }
    if (this.root && schema) {
      var validationResult = new ValidationResult();
      validate(this.root, schema, validationResult, NoOpSchemaCollector.instance);
      return validationResult.problems.map(function(p) {
        var _a;
        var range = Range.create(textDocument.positionAt(p.location.offset), textDocument.positionAt(p.location.offset + p.location.length));
        return Diagnostic.create(range, p.message, (_a = p.severity) !== null && _a !== void 0 ? _a : severity, p.code);
      });
    }
    return void 0;
  };
  JSONDocument3.prototype.getMatchingSchemas = function(schema, focusOffset, exclude) {
    if (focusOffset === void 0) {
      focusOffset = -1;
    }
    var matchingSchemas = new SchemaCollector(focusOffset, exclude);
    if (this.root && schema) {
      validate(this.root, schema, new ValidationResult(), matchingSchemas);
    }
    return matchingSchemas.schemas;
  };
  return JSONDocument3;
}();
function validate(n, schema, validationResult, matchingSchemas) {
  if (!n || !matchingSchemas.include(n)) {
    return;
  }
  var node = n;
  switch (node.type) {
    case "object":
      _validateObjectNode(node, schema, validationResult, matchingSchemas);
      break;
    case "array":
      _validateArrayNode(node, schema, validationResult, matchingSchemas);
      break;
    case "string":
      _validateStringNode(node, schema, validationResult, matchingSchemas);
      break;
    case "number":
      _validateNumberNode(node, schema, validationResult, matchingSchemas);
      break;
    case "property":
      return validate(node.valueNode, schema, validationResult, matchingSchemas);
  }
  _validateNode();
  matchingSchemas.add({ node, schema });
  function _validateNode() {
    function matchesType(type) {
      return node.type === type || type === "integer" && node.type === "number" && node.isInteger;
    }
    if (Array.isArray(schema.type)) {
      if (!schema.type.some(matchesType)) {
        validationResult.problems.push({
          location: { offset: node.offset, length: node.length },
          message: schema.errorMessage || localize2("typeArrayMismatchWarning", "Incorrect type. Expected one of {0}.", schema.type.join(", "))
        });
      }
    } else if (schema.type) {
      if (!matchesType(schema.type)) {
        validationResult.problems.push({
          location: { offset: node.offset, length: node.length },
          message: schema.errorMessage || localize2("typeMismatchWarning", 'Incorrect type. Expected "{0}".', schema.type)
        });
      }
    }
    if (Array.isArray(schema.allOf)) {
      for (var _i = 0, _a = schema.allOf; _i < _a.length; _i++) {
        var subSchemaRef = _a[_i];
        validate(node, asSchema(subSchemaRef), validationResult, matchingSchemas);
      }
    }
    var notSchema = asSchema(schema.not);
    if (notSchema) {
      var subValidationResult = new ValidationResult();
      var subMatchingSchemas = matchingSchemas.newSub();
      validate(node, notSchema, subValidationResult, subMatchingSchemas);
      if (!subValidationResult.hasProblems()) {
        validationResult.problems.push({
          location: { offset: node.offset, length: node.length },
          message: localize2("notSchemaWarning", "Matches a schema that is not allowed.")
        });
      }
      for (var _b = 0, _c = subMatchingSchemas.schemas; _b < _c.length; _b++) {
        var ms = _c[_b];
        ms.inverted = !ms.inverted;
        matchingSchemas.add(ms);
      }
    }
    var testAlternatives = function(alternatives, maxOneMatch) {
      var matches = [];
      var bestMatch = void 0;
      for (var _i2 = 0, alternatives_1 = alternatives; _i2 < alternatives_1.length; _i2++) {
        var subSchemaRef2 = alternatives_1[_i2];
        var subSchema = asSchema(subSchemaRef2);
        var subValidationResult2 = new ValidationResult();
        var subMatchingSchemas2 = matchingSchemas.newSub();
        validate(node, subSchema, subValidationResult2, subMatchingSchemas2);
        if (!subValidationResult2.hasProblems()) {
          matches.push(subSchema);
        }
        if (!bestMatch) {
          bestMatch = { schema: subSchema, validationResult: subValidationResult2, matchingSchemas: subMatchingSchemas2 };
        } else {
          if (!maxOneMatch && !subValidationResult2.hasProblems() && !bestMatch.validationResult.hasProblems()) {
            bestMatch.matchingSchemas.merge(subMatchingSchemas2);
            bestMatch.validationResult.propertiesMatches += subValidationResult2.propertiesMatches;
            bestMatch.validationResult.propertiesValueMatches += subValidationResult2.propertiesValueMatches;
          } else {
            var compareResult = subValidationResult2.compare(bestMatch.validationResult);
            if (compareResult > 0) {
              bestMatch = { schema: subSchema, validationResult: subValidationResult2, matchingSchemas: subMatchingSchemas2 };
            } else if (compareResult === 0) {
              bestMatch.matchingSchemas.merge(subMatchingSchemas2);
              bestMatch.validationResult.mergeEnumValues(subValidationResult2);
            }
          }
        }
      }
      if (matches.length > 1 && maxOneMatch) {
        validationResult.problems.push({
          location: { offset: node.offset, length: 1 },
          message: localize2("oneOfWarning", "Matches multiple schemas when only one must validate.")
        });
      }
      if (bestMatch) {
        validationResult.merge(bestMatch.validationResult);
        validationResult.propertiesMatches += bestMatch.validationResult.propertiesMatches;
        validationResult.propertiesValueMatches += bestMatch.validationResult.propertiesValueMatches;
        matchingSchemas.merge(bestMatch.matchingSchemas);
      }
      return matches.length;
    };
    if (Array.isArray(schema.anyOf)) {
      testAlternatives(schema.anyOf, false);
    }
    if (Array.isArray(schema.oneOf)) {
      testAlternatives(schema.oneOf, true);
    }
    var testBranch = function(schema2) {
      var subValidationResult2 = new ValidationResult();
      var subMatchingSchemas2 = matchingSchemas.newSub();
      validate(node, asSchema(schema2), subValidationResult2, subMatchingSchemas2);
      validationResult.merge(subValidationResult2);
      validationResult.propertiesMatches += subValidationResult2.propertiesMatches;
      validationResult.propertiesValueMatches += subValidationResult2.propertiesValueMatches;
      matchingSchemas.merge(subMatchingSchemas2);
    };
    var testCondition = function(ifSchema2, thenSchema, elseSchema) {
      var subSchema = asSchema(ifSchema2);
      var subValidationResult2 = new ValidationResult();
      var subMatchingSchemas2 = matchingSchemas.newSub();
      validate(node, subSchema, subValidationResult2, subMatchingSchemas2);
      matchingSchemas.merge(subMatchingSchemas2);
      if (!subValidationResult2.hasProblems()) {
        if (thenSchema) {
          testBranch(thenSchema);
        }
      } else if (elseSchema) {
        testBranch(elseSchema);
      }
    };
    var ifSchema = asSchema(schema.if);
    if (ifSchema) {
      testCondition(ifSchema, asSchema(schema.then), asSchema(schema.else));
    }
    if (Array.isArray(schema.enum)) {
      var val = getNodeValue2(node);
      var enumValueMatch = false;
      for (var _d = 0, _e = schema.enum; _d < _e.length; _d++) {
        var e = _e[_d];
        if (equals(val, e)) {
          enumValueMatch = true;
          break;
        }
      }
      validationResult.enumValues = schema.enum;
      validationResult.enumValueMatch = enumValueMatch;
      if (!enumValueMatch) {
        validationResult.problems.push({
          location: { offset: node.offset, length: node.length },
          code: ErrorCode.EnumValueMismatch,
          message: schema.errorMessage || localize2("enumWarning", "Value is not accepted. Valid values: {0}.", schema.enum.map(function(v) {
            return JSON.stringify(v);
          }).join(", "))
        });
      }
    }
    if (isDefined(schema.const)) {
      var val = getNodeValue2(node);
      if (!equals(val, schema.const)) {
        validationResult.problems.push({
          location: { offset: node.offset, length: node.length },
          code: ErrorCode.EnumValueMismatch,
          message: schema.errorMessage || localize2("constWarning", "Value must be {0}.", JSON.stringify(schema.const))
        });
        validationResult.enumValueMatch = false;
      } else {
        validationResult.enumValueMatch = true;
      }
      validationResult.enumValues = [schema.const];
    }
    if (schema.deprecationMessage && node.parent) {
      validationResult.problems.push({
        location: { offset: node.parent.offset, length: node.parent.length },
        severity: DiagnosticSeverity.Warning,
        message: schema.deprecationMessage,
        code: ErrorCode.Deprecated
      });
    }
  }
  function _validateNumberNode(node2, schema2, validationResult2, matchingSchemas2) {
    var val = node2.value;
    function normalizeFloats(float) {
      var _a;
      var parts = /^(-?\d+)(?:\.(\d+))?(?:e([-+]\d+))?$/.exec(float.toString());
      return parts && {
        value: Number(parts[1] + (parts[2] || "")),
        multiplier: (((_a = parts[2]) === null || _a === void 0 ? void 0 : _a.length) || 0) - (parseInt(parts[3]) || 0)
      };
    }
    ;
    if (isNumber(schema2.multipleOf)) {
      var remainder = -1;
      if (Number.isInteger(schema2.multipleOf)) {
        remainder = val % schema2.multipleOf;
      } else {
        var normMultipleOf = normalizeFloats(schema2.multipleOf);
        var normValue = normalizeFloats(val);
        if (normMultipleOf && normValue) {
          var multiplier = Math.pow(10, Math.abs(normValue.multiplier - normMultipleOf.multiplier));
          if (normValue.multiplier < normMultipleOf.multiplier) {
            normValue.value *= multiplier;
          } else {
            normMultipleOf.value *= multiplier;
          }
          remainder = normValue.value % normMultipleOf.value;
        }
      }
      if (remainder !== 0) {
        validationResult2.problems.push({
          location: { offset: node2.offset, length: node2.length },
          message: localize2("multipleOfWarning", "Value is not divisible by {0}.", schema2.multipleOf)
        });
      }
    }
    function getExclusiveLimit(limit, exclusive) {
      if (isNumber(exclusive)) {
        return exclusive;
      }
      if (isBoolean(exclusive) && exclusive) {
        return limit;
      }
      return void 0;
    }
    function getLimit(limit, exclusive) {
      if (!isBoolean(exclusive) || !exclusive) {
        return limit;
      }
      return void 0;
    }
    var exclusiveMinimum = getExclusiveLimit(schema2.minimum, schema2.exclusiveMinimum);
    if (isNumber(exclusiveMinimum) && val <= exclusiveMinimum) {
      validationResult2.problems.push({
        location: { offset: node2.offset, length: node2.length },
        message: localize2("exclusiveMinimumWarning", "Value is below the exclusive minimum of {0}.", exclusiveMinimum)
      });
    }
    var exclusiveMaximum = getExclusiveLimit(schema2.maximum, schema2.exclusiveMaximum);
    if (isNumber(exclusiveMaximum) && val >= exclusiveMaximum) {
      validationResult2.problems.push({
        location: { offset: node2.offset, length: node2.length },
        message: localize2("exclusiveMaximumWarning", "Value is above the exclusive maximum of {0}.", exclusiveMaximum)
      });
    }
    var minimum = getLimit(schema2.minimum, schema2.exclusiveMinimum);
    if (isNumber(minimum) && val < minimum) {
      validationResult2.problems.push({
        location: { offset: node2.offset, length: node2.length },
        message: localize2("minimumWarning", "Value is below the minimum of {0}.", minimum)
      });
    }
    var maximum = getLimit(schema2.maximum, schema2.exclusiveMaximum);
    if (isNumber(maximum) && val > maximum) {
      validationResult2.problems.push({
        location: { offset: node2.offset, length: node2.length },
        message: localize2("maximumWarning", "Value is above the maximum of {0}.", maximum)
      });
    }
  }
  function _validateStringNode(node2, schema2, validationResult2, matchingSchemas2) {
    if (isNumber(schema2.minLength) && node2.value.length < schema2.minLength) {
      validationResult2.problems.push({
        location: { offset: node2.offset, length: node2.length },
        message: localize2("minLengthWarning", "String is shorter than the minimum length of {0}.", schema2.minLength)
      });
    }
    if (isNumber(schema2.maxLength) && node2.value.length > schema2.maxLength) {
      validationResult2.problems.push({
        location: { offset: node2.offset, length: node2.length },
        message: localize2("maxLengthWarning", "String is longer than the maximum length of {0}.", schema2.maxLength)
      });
    }
    if (isString(schema2.pattern)) {
      var regex = extendedRegExp(schema2.pattern);
      if (!(regex === null || regex === void 0 ? void 0 : regex.test(node2.value))) {
        validationResult2.problems.push({
          location: { offset: node2.offset, length: node2.length },
          message: schema2.patternErrorMessage || schema2.errorMessage || localize2("patternWarning", 'String does not match the pattern of "{0}".', schema2.pattern)
        });
      }
    }
    if (schema2.format) {
      switch (schema2.format) {
        case "uri":
        case "uri-reference":
          {
            var errorMessage = void 0;
            if (!node2.value) {
              errorMessage = localize2("uriEmpty", "URI expected.");
            } else {
              var match = /^(([^:/?#]+?):)?(\/\/([^/?#]*))?([^?#]*)(\?([^#]*))?(#(.*))?/.exec(node2.value);
              if (!match) {
                errorMessage = localize2("uriMissing", "URI is expected.");
              } else if (!match[2] && schema2.format === "uri") {
                errorMessage = localize2("uriSchemeMissing", "URI with a scheme is expected.");
              }
            }
            if (errorMessage) {
              validationResult2.problems.push({
                location: { offset: node2.offset, length: node2.length },
                message: schema2.patternErrorMessage || schema2.errorMessage || localize2("uriFormatWarning", "String is not a URI: {0}", errorMessage)
              });
            }
          }
          break;
        case "color-hex":
        case "date-time":
        case "date":
        case "time":
        case "email":
          var format3 = formats[schema2.format];
          if (!node2.value || !format3.pattern.exec(node2.value)) {
            validationResult2.problems.push({
              location: { offset: node2.offset, length: node2.length },
              message: schema2.patternErrorMessage || schema2.errorMessage || format3.errorMessage
            });
          }
        default:
      }
    }
  }
  function _validateArrayNode(node2, schema2, validationResult2, matchingSchemas2) {
    if (Array.isArray(schema2.items)) {
      var subSchemas = schema2.items;
      for (var index = 0; index < subSchemas.length; index++) {
        var subSchemaRef = subSchemas[index];
        var subSchema = asSchema(subSchemaRef);
        var itemValidationResult = new ValidationResult();
        var item = node2.items[index];
        if (item) {
          validate(item, subSchema, itemValidationResult, matchingSchemas2);
          validationResult2.mergePropertyMatch(itemValidationResult);
        } else if (node2.items.length >= subSchemas.length) {
          validationResult2.propertiesValueMatches++;
        }
      }
      if (node2.items.length > subSchemas.length) {
        if (typeof schema2.additionalItems === "object") {
          for (var i = subSchemas.length; i < node2.items.length; i++) {
            var itemValidationResult = new ValidationResult();
            validate(node2.items[i], schema2.additionalItems, itemValidationResult, matchingSchemas2);
            validationResult2.mergePropertyMatch(itemValidationResult);
          }
        } else if (schema2.additionalItems === false) {
          validationResult2.problems.push({
            location: { offset: node2.offset, length: node2.length },
            message: localize2("additionalItemsWarning", "Array has too many items according to schema. Expected {0} or fewer.", subSchemas.length)
          });
        }
      }
    } else {
      var itemSchema = asSchema(schema2.items);
      if (itemSchema) {
        for (var _i = 0, _a = node2.items; _i < _a.length; _i++) {
          var item = _a[_i];
          var itemValidationResult = new ValidationResult();
          validate(item, itemSchema, itemValidationResult, matchingSchemas2);
          validationResult2.mergePropertyMatch(itemValidationResult);
        }
      }
    }
    var containsSchema = asSchema(schema2.contains);
    if (containsSchema) {
      var doesContain = node2.items.some(function(item2) {
        var itemValidationResult2 = new ValidationResult();
        validate(item2, containsSchema, itemValidationResult2, NoOpSchemaCollector.instance);
        return !itemValidationResult2.hasProblems();
      });
      if (!doesContain) {
        validationResult2.problems.push({
          location: { offset: node2.offset, length: node2.length },
          message: schema2.errorMessage || localize2("requiredItemMissingWarning", "Array does not contain required item.")
        });
      }
    }
    if (isNumber(schema2.minItems) && node2.items.length < schema2.minItems) {
      validationResult2.problems.push({
        location: { offset: node2.offset, length: node2.length },
        message: localize2("minItemsWarning", "Array has too few items. Expected {0} or more.", schema2.minItems)
      });
    }
    if (isNumber(schema2.maxItems) && node2.items.length > schema2.maxItems) {
      validationResult2.problems.push({
        location: { offset: node2.offset, length: node2.length },
        message: localize2("maxItemsWarning", "Array has too many items. Expected {0} or fewer.", schema2.maxItems)
      });
    }
    if (schema2.uniqueItems === true) {
      var values_1 = getNodeValue2(node2);
      var duplicates = values_1.some(function(value, index2) {
        return index2 !== values_1.lastIndexOf(value);
      });
      if (duplicates) {
        validationResult2.problems.push({
          location: { offset: node2.offset, length: node2.length },
          message: localize2("uniqueItemsWarning", "Array has duplicate items.")
        });
      }
    }
  }
  function _validateObjectNode(node2, schema2, validationResult2, matchingSchemas2) {
    var seenKeys = Object.create(null);
    var unprocessedProperties = [];
    for (var _i = 0, _a = node2.properties; _i < _a.length; _i++) {
      var propertyNode = _a[_i];
      var key = propertyNode.keyNode.value;
      seenKeys[key] = propertyNode.valueNode;
      unprocessedProperties.push(key);
    }
    if (Array.isArray(schema2.required)) {
      for (var _b = 0, _c = schema2.required; _b < _c.length; _b++) {
        var propertyName = _c[_b];
        if (!seenKeys[propertyName]) {
          var keyNode = node2.parent && node2.parent.type === "property" && node2.parent.keyNode;
          var location = keyNode ? { offset: keyNode.offset, length: keyNode.length } : { offset: node2.offset, length: 1 };
          validationResult2.problems.push({
            location,
            message: localize2("MissingRequiredPropWarning", 'Missing property "{0}".', propertyName)
          });
        }
      }
    }
    var propertyProcessed = function(prop2) {
      var index = unprocessedProperties.indexOf(prop2);
      while (index >= 0) {
        unprocessedProperties.splice(index, 1);
        index = unprocessedProperties.indexOf(prop2);
      }
    };
    if (schema2.properties) {
      for (var _d = 0, _e = Object.keys(schema2.properties); _d < _e.length; _d++) {
        var propertyName = _e[_d];
        propertyProcessed(propertyName);
        var propertySchema = schema2.properties[propertyName];
        var child = seenKeys[propertyName];
        if (child) {
          if (isBoolean(propertySchema)) {
            if (!propertySchema) {
              var propertyNode = child.parent;
              validationResult2.problems.push({
                location: { offset: propertyNode.keyNode.offset, length: propertyNode.keyNode.length },
                message: schema2.errorMessage || localize2("DisallowedExtraPropWarning", "Property {0} is not allowed.", propertyName)
              });
            } else {
              validationResult2.propertiesMatches++;
              validationResult2.propertiesValueMatches++;
            }
          } else {
            var propertyValidationResult = new ValidationResult();
            validate(child, propertySchema, propertyValidationResult, matchingSchemas2);
            validationResult2.mergePropertyMatch(propertyValidationResult);
          }
        }
      }
    }
    if (schema2.patternProperties) {
      for (var _f = 0, _g = Object.keys(schema2.patternProperties); _f < _g.length; _f++) {
        var propertyPattern = _g[_f];
        var regex = extendedRegExp(propertyPattern);
        for (var _h = 0, _j = unprocessedProperties.slice(0); _h < _j.length; _h++) {
          var propertyName = _j[_h];
          if (regex === null || regex === void 0 ? void 0 : regex.test(propertyName)) {
            propertyProcessed(propertyName);
            var child = seenKeys[propertyName];
            if (child) {
              var propertySchema = schema2.patternProperties[propertyPattern];
              if (isBoolean(propertySchema)) {
                if (!propertySchema) {
                  var propertyNode = child.parent;
                  validationResult2.problems.push({
                    location: { offset: propertyNode.keyNode.offset, length: propertyNode.keyNode.length },
                    message: schema2.errorMessage || localize2("DisallowedExtraPropWarning", "Property {0} is not allowed.", propertyName)
                  });
                } else {
                  validationResult2.propertiesMatches++;
                  validationResult2.propertiesValueMatches++;
                }
              } else {
                var propertyValidationResult = new ValidationResult();
                validate(child, propertySchema, propertyValidationResult, matchingSchemas2);
                validationResult2.mergePropertyMatch(propertyValidationResult);
              }
            }
          }
        }
      }
    }
    if (typeof schema2.additionalProperties === "object") {
      for (var _k = 0, unprocessedProperties_1 = unprocessedProperties; _k < unprocessedProperties_1.length; _k++) {
        var propertyName = unprocessedProperties_1[_k];
        var child = seenKeys[propertyName];
        if (child) {
          var propertyValidationResult = new ValidationResult();
          validate(child, schema2.additionalProperties, propertyValidationResult, matchingSchemas2);
          validationResult2.mergePropertyMatch(propertyValidationResult);
        }
      }
    } else if (schema2.additionalProperties === false) {
      if (unprocessedProperties.length > 0) {
        for (var _l = 0, unprocessedProperties_2 = unprocessedProperties; _l < unprocessedProperties_2.length; _l++) {
          var propertyName = unprocessedProperties_2[_l];
          var child = seenKeys[propertyName];
          if (child) {
            var propertyNode = child.parent;
            validationResult2.problems.push({
              location: { offset: propertyNode.keyNode.offset, length: propertyNode.keyNode.length },
              message: schema2.errorMessage || localize2("DisallowedExtraPropWarning", "Property {0} is not allowed.", propertyName)
            });
          }
        }
      }
    }
    if (isNumber(schema2.maxProperties)) {
      if (node2.properties.length > schema2.maxProperties) {
        validationResult2.problems.push({
          location: { offset: node2.offset, length: node2.length },
          message: localize2("MaxPropWarning", "Object has more properties than limit of {0}.", schema2.maxProperties)
        });
      }
    }
    if (isNumber(schema2.minProperties)) {
      if (node2.properties.length < schema2.minProperties) {
        validationResult2.problems.push({
          location: { offset: node2.offset, length: node2.length },
          message: localize2("MinPropWarning", "Object has fewer properties than the required number of {0}", schema2.minProperties)
        });
      }
    }
    if (schema2.dependencies) {
      for (var _m = 0, _o = Object.keys(schema2.dependencies); _m < _o.length; _m++) {
        var key = _o[_m];
        var prop = seenKeys[key];
        if (prop) {
          var propertyDep = schema2.dependencies[key];
          if (Array.isArray(propertyDep)) {
            for (var _p = 0, propertyDep_1 = propertyDep; _p < propertyDep_1.length; _p++) {
              var requiredProp = propertyDep_1[_p];
              if (!seenKeys[requiredProp]) {
                validationResult2.problems.push({
                  location: { offset: node2.offset, length: node2.length },
                  message: localize2("RequiredDependentPropWarning", "Object is missing property {0} required by property {1}.", requiredProp, key)
                });
              } else {
                validationResult2.propertiesValueMatches++;
              }
            }
          } else {
            var propertySchema = asSchema(propertyDep);
            if (propertySchema) {
              var propertyValidationResult = new ValidationResult();
              validate(node2, propertySchema, propertyValidationResult, matchingSchemas2);
              validationResult2.mergePropertyMatch(propertyValidationResult);
            }
          }
        }
      }
    }
    var propertyNames = asSchema(schema2.propertyNames);
    if (propertyNames) {
      for (var _q = 0, _r = node2.properties; _q < _r.length; _q++) {
        var f2 = _r[_q];
        var key = f2.keyNode;
        if (key) {
          validate(key, propertyNames, validationResult2, NoOpSchemaCollector.instance);
        }
      }
    }
  }
}

// node_modules/vscode-json-languageservice/lib/esm/utils/glob.js
function createRegex(glob, opts) {
  if (typeof glob !== "string") {
    throw new TypeError("Expected a string");
  }
  var str = String(glob);
  var reStr = "";
  var extended = opts ? !!opts.extended : false;
  var globstar = opts ? !!opts.globstar : false;
  var inGroup = false;
  var flags = opts && typeof opts.flags === "string" ? opts.flags : "";
  var c;
  for (var i = 0, len = str.length; i < len; i++) {
    c = str[i];
    switch (c) {
      case "/":
      case "$":
      case "^":
      case "+":
      case ".":
      case "(":
      case ")":
      case "=":
      case "!":
      case "|":
        reStr += "\\" + c;
        break;
      case "?":
        if (extended) {
          reStr += ".";
          break;
        }
      case "[":
      case "]":
        if (extended) {
          reStr += c;
          break;
        }
      case "{":
        if (extended) {
          inGroup = true;
          reStr += "(";
          break;
        }
      case "}":
        if (extended) {
          inGroup = false;
          reStr += ")";
          break;
        }
      case ",":
        if (inGroup) {
          reStr += "|";
          break;
        }
        reStr += "\\" + c;
        break;
      case "*":
        var prevChar = str[i - 1];
        var starCount = 1;
        while (str[i + 1] === "*") {
          starCount++;
          i++;
        }
        var nextChar = str[i + 1];
        if (!globstar) {
          reStr += ".*";
        } else {
          var isGlobstar = starCount > 1 && (prevChar === "/" || prevChar === void 0 || prevChar === "{" || prevChar === ",") && (nextChar === "/" || nextChar === void 0 || nextChar === "," || nextChar === "}");
          if (isGlobstar) {
            if (nextChar === "/") {
              i++;
            } else if (prevChar === "/" && reStr.endsWith("\\/")) {
              reStr = reStr.substr(0, reStr.length - 2);
            }
            reStr += "((?:[^/]*(?:/|$))*)";
          } else {
            reStr += "([^/]*)";
          }
        }
        break;
      default:
        reStr += c;
    }
  }
  if (!flags || !~flags.indexOf("g")) {
    reStr = "^" + reStr + "$";
  }
  return new RegExp(reStr, flags);
}

// node_modules/vscode-json-languageservice/lib/esm/services/jsonSchemaService.js
var localize3 = loadMessageBundle();
var BANG = "!";
var PATH_SEP = "/";
var FilePatternAssociation = function() {
  function FilePatternAssociation3(pattern, uris) {
    this.globWrappers = [];
    try {
      for (var _i = 0, pattern_1 = pattern; _i < pattern_1.length; _i++) {
        var patternString = pattern_1[_i];
        var include = patternString[0] !== BANG;
        if (!include) {
          patternString = patternString.substring(1);
        }
        if (patternString.length > 0) {
          if (patternString[0] === PATH_SEP) {
            patternString = patternString.substring(1);
          }
          this.globWrappers.push({
            regexp: createRegex("**/" + patternString, { extended: true, globstar: true }),
            include
          });
        }
      }
      ;
      this.uris = uris;
    } catch (e) {
      this.globWrappers.length = 0;
      this.uris = [];
    }
  }
  FilePatternAssociation3.prototype.matchesPattern = function(fileName) {
    var match = false;
    for (var _i = 0, _a = this.globWrappers; _i < _a.length; _i++) {
      var _b = _a[_i], regexp = _b.regexp, include = _b.include;
      if (regexp.test(fileName)) {
        match = include;
      }
    }
    return match;
  };
  FilePatternAssociation3.prototype.getURIs = function() {
    return this.uris;
  };
  return FilePatternAssociation3;
}();
var SchemaHandle = function() {
  function SchemaHandle2(service, url, unresolvedSchemaContent) {
    this.service = service;
    this.url = url;
    this.dependencies = {};
    if (unresolvedSchemaContent) {
      this.unresolvedSchema = this.service.promise.resolve(new UnresolvedSchema(unresolvedSchemaContent));
    }
  }
  SchemaHandle2.prototype.getUnresolvedSchema = function() {
    if (!this.unresolvedSchema) {
      this.unresolvedSchema = this.service.loadSchema(this.url);
    }
    return this.unresolvedSchema;
  };
  SchemaHandle2.prototype.getResolvedSchema = function() {
    var _this = this;
    if (!this.resolvedSchema) {
      this.resolvedSchema = this.getUnresolvedSchema().then(function(unresolved) {
        return _this.service.resolveSchemaContent(unresolved, _this.url, _this.dependencies);
      });
    }
    return this.resolvedSchema;
  };
  SchemaHandle2.prototype.clearSchema = function() {
    this.resolvedSchema = void 0;
    this.unresolvedSchema = void 0;
    this.dependencies = {};
  };
  return SchemaHandle2;
}();
var UnresolvedSchema = function() {
  function UnresolvedSchema2(schema, errors) {
    if (errors === void 0) {
      errors = [];
    }
    this.schema = schema;
    this.errors = errors;
  }
  return UnresolvedSchema2;
}();
var ResolvedSchema = function() {
  function ResolvedSchema2(schema, errors) {
    if (errors === void 0) {
      errors = [];
    }
    this.schema = schema;
    this.errors = errors;
  }
  ResolvedSchema2.prototype.getSection = function(path5) {
    var schemaRef = this.getSectionRecursive(path5, this.schema);
    if (schemaRef) {
      return asSchema(schemaRef);
    }
    return void 0;
  };
  ResolvedSchema2.prototype.getSectionRecursive = function(path5, schema) {
    if (!schema || typeof schema === "boolean" || path5.length === 0) {
      return schema;
    }
    var next = path5.shift();
    if (schema.properties && typeof schema.properties[next]) {
      return this.getSectionRecursive(path5, schema.properties[next]);
    } else if (schema.patternProperties) {
      for (var _i = 0, _a = Object.keys(schema.patternProperties); _i < _a.length; _i++) {
        var pattern = _a[_i];
        var regex = extendedRegExp(pattern);
        if (regex === null || regex === void 0 ? void 0 : regex.test(next)) {
          return this.getSectionRecursive(path5, schema.patternProperties[pattern]);
        }
      }
    } else if (typeof schema.additionalProperties === "object") {
      return this.getSectionRecursive(path5, schema.additionalProperties);
    } else if (next.match("[0-9]+")) {
      if (Array.isArray(schema.items)) {
        var index = parseInt(next, 10);
        if (!isNaN(index) && schema.items[index]) {
          return this.getSectionRecursive(path5, schema.items[index]);
        }
      } else if (schema.items) {
        return this.getSectionRecursive(path5, schema.items);
      }
    }
    return void 0;
  };
  return ResolvedSchema2;
}();
var JSONSchemaService = function() {
  function JSONSchemaService2(requestService, contextService, promiseConstructor) {
    this.contextService = contextService;
    this.requestService = requestService;
    this.promiseConstructor = promiseConstructor || Promise;
    this.callOnDispose = [];
    this.contributionSchemas = {};
    this.contributionAssociations = [];
    this.schemasById = {};
    this.filePatternAssociations = [];
    this.registeredSchemasIds = {};
  }
  JSONSchemaService2.prototype.getRegisteredSchemaIds = function(filter) {
    return Object.keys(this.registeredSchemasIds).filter(function(id) {
      var scheme = URI.parse(id).scheme;
      return scheme !== "schemaservice" && (!filter || filter(scheme));
    });
  };
  Object.defineProperty(JSONSchemaService2.prototype, "promise", {
    get: function() {
      return this.promiseConstructor;
    },
    enumerable: false,
    configurable: true
  });
  JSONSchemaService2.prototype.dispose = function() {
    while (this.callOnDispose.length > 0) {
      this.callOnDispose.pop()();
    }
  };
  JSONSchemaService2.prototype.onResourceChange = function(uri) {
    var _this = this;
    this.cachedSchemaForResource = void 0;
    var hasChanges = false;
    uri = normalizeId(uri);
    var toWalk = [uri];
    var all = Object.keys(this.schemasById).map(function(key) {
      return _this.schemasById[key];
    });
    while (toWalk.length) {
      var curr = toWalk.pop();
      for (var i = 0; i < all.length; i++) {
        var handle = all[i];
        if (handle && (handle.url === curr || handle.dependencies[curr])) {
          if (handle.url !== curr) {
            toWalk.push(handle.url);
          }
          handle.clearSchema();
          all[i] = void 0;
          hasChanges = true;
        }
      }
    }
    return hasChanges;
  };
  JSONSchemaService2.prototype.setSchemaContributions = function(schemaContributions2) {
    if (schemaContributions2.schemas) {
      var schemas = schemaContributions2.schemas;
      for (var id in schemas) {
        var normalizedId = normalizeId(id);
        this.contributionSchemas[normalizedId] = this.addSchemaHandle(normalizedId, schemas[id]);
      }
    }
    if (Array.isArray(schemaContributions2.schemaAssociations)) {
      var schemaAssociations = schemaContributions2.schemaAssociations;
      for (var _i = 0, schemaAssociations_1 = schemaAssociations; _i < schemaAssociations_1.length; _i++) {
        var schemaAssociation = schemaAssociations_1[_i];
        var uris = schemaAssociation.uris.map(normalizeId);
        var association = this.addFilePatternAssociation(schemaAssociation.pattern, uris);
        this.contributionAssociations.push(association);
      }
    }
  };
  JSONSchemaService2.prototype.addSchemaHandle = function(id, unresolvedSchemaContent) {
    var schemaHandle = new SchemaHandle(this, id, unresolvedSchemaContent);
    this.schemasById[id] = schemaHandle;
    return schemaHandle;
  };
  JSONSchemaService2.prototype.getOrAddSchemaHandle = function(id, unresolvedSchemaContent) {
    return this.schemasById[id] || this.addSchemaHandle(id, unresolvedSchemaContent);
  };
  JSONSchemaService2.prototype.addFilePatternAssociation = function(pattern, uris) {
    var fpa = new FilePatternAssociation(pattern, uris);
    this.filePatternAssociations.push(fpa);
    return fpa;
  };
  JSONSchemaService2.prototype.registerExternalSchema = function(uri, filePatterns, unresolvedSchemaContent) {
    var id = normalizeId(uri);
    this.registeredSchemasIds[id] = true;
    this.cachedSchemaForResource = void 0;
    if (filePatterns) {
      this.addFilePatternAssociation(filePatterns, [id]);
    }
    return unresolvedSchemaContent ? this.addSchemaHandle(id, unresolvedSchemaContent) : this.getOrAddSchemaHandle(id);
  };
  JSONSchemaService2.prototype.clearExternalSchemas = function() {
    this.schemasById = {};
    this.filePatternAssociations = [];
    this.registeredSchemasIds = {};
    this.cachedSchemaForResource = void 0;
    for (var id in this.contributionSchemas) {
      this.schemasById[id] = this.contributionSchemas[id];
      this.registeredSchemasIds[id] = true;
    }
    for (var _i = 0, _a = this.contributionAssociations; _i < _a.length; _i++) {
      var contributionAssociation = _a[_i];
      this.filePatternAssociations.push(contributionAssociation);
    }
  };
  JSONSchemaService2.prototype.getResolvedSchema = function(schemaId) {
    var id = normalizeId(schemaId);
    var schemaHandle = this.schemasById[id];
    if (schemaHandle) {
      return schemaHandle.getResolvedSchema();
    }
    return this.promise.resolve(void 0);
  };
  JSONSchemaService2.prototype.loadSchema = function(url) {
    if (!this.requestService) {
      var errorMessage = localize3("json.schema.norequestservice", "Unable to load schema from '{0}'. No schema request service available", toDisplayString(url));
      return this.promise.resolve(new UnresolvedSchema({}, [errorMessage]));
    }
    return this.requestService(url).then(function(content) {
      if (!content) {
        var errorMessage2 = localize3("json.schema.nocontent", "Unable to load schema from '{0}': No content.", toDisplayString(url));
        return new UnresolvedSchema({}, [errorMessage2]);
      }
      var schemaContent = {};
      var jsonErrors = [];
      schemaContent = parse(content, jsonErrors);
      var errors = jsonErrors.length ? [localize3("json.schema.invalidFormat", "Unable to parse content from '{0}': Parse error at offset {1}.", toDisplayString(url), jsonErrors[0].offset)] : [];
      return new UnresolvedSchema(schemaContent, errors);
    }, function(error) {
      var errorMessage2 = error.toString();
      var errorSplit = error.toString().split("Error: ");
      if (errorSplit.length > 1) {
        errorMessage2 = errorSplit[1];
      }
      if (endsWith(errorMessage2, ".")) {
        errorMessage2 = errorMessage2.substr(0, errorMessage2.length - 1);
      }
      return new UnresolvedSchema({}, [localize3("json.schema.nocontent", "Unable to load schema from '{0}': {1}.", toDisplayString(url), errorMessage2)]);
    });
  };
  JSONSchemaService2.prototype.resolveSchemaContent = function(schemaToResolve, schemaURL, dependencies) {
    var _this = this;
    var resolveErrors = schemaToResolve.errors.slice(0);
    var schema = schemaToResolve.schema;
    if (schema.$schema) {
      var id = normalizeId(schema.$schema);
      if (id === "http://json-schema.org/draft-03/schema") {
        return this.promise.resolve(new ResolvedSchema({}, [localize3("json.schema.draft03.notsupported", "Draft-03 schemas are not supported.")]));
      } else if (id === "https://json-schema.org/draft/2019-09/schema") {
        resolveErrors.push(localize3("json.schema.draft201909.notsupported", "Draft 2019-09 schemas are not yet fully supported."));
      }
    }
    var contextService = this.contextService;
    var findSection = function(schema2, path5) {
      if (!path5) {
        return schema2;
      }
      var current = schema2;
      if (path5[0] === "/") {
        path5 = path5.substr(1);
      }
      path5.split("/").some(function(part) {
        part = part.replace(/~1/g, "/").replace(/~0/g, "~");
        current = current[part];
        return !current;
      });
      return current;
    };
    var merge = function(target, sourceRoot, sourceURI, refSegment) {
      var path5 = refSegment ? decodeURIComponent(refSegment) : void 0;
      var section = findSection(sourceRoot, path5);
      if (section) {
        for (var key in section) {
          if (section.hasOwnProperty(key) && !target.hasOwnProperty(key)) {
            target[key] = section[key];
          }
        }
      } else {
        resolveErrors.push(localize3("json.schema.invalidref", "$ref '{0}' in '{1}' can not be resolved.", path5, sourceURI));
      }
    };
    var resolveExternalLink = function(node, uri, refSegment, parentSchemaURL, parentSchemaDependencies) {
      if (contextService && !/^[A-Za-z][A-Za-z0-9+\-.+]*:\/\/.*/.test(uri)) {
        uri = contextService.resolveRelativePath(uri, parentSchemaURL);
      }
      uri = normalizeId(uri);
      var referencedHandle = _this.getOrAddSchemaHandle(uri);
      return referencedHandle.getUnresolvedSchema().then(function(unresolvedSchema) {
        parentSchemaDependencies[uri] = true;
        if (unresolvedSchema.errors.length) {
          var loc = refSegment ? uri + "#" + refSegment : uri;
          resolveErrors.push(localize3("json.schema.problemloadingref", "Problems loading reference '{0}': {1}", loc, unresolvedSchema.errors[0]));
        }
        merge(node, unresolvedSchema.schema, uri, refSegment);
        return resolveRefs(node, unresolvedSchema.schema, uri, referencedHandle.dependencies);
      });
    };
    var resolveRefs = function(node, parentSchema, parentSchemaURL, parentSchemaDependencies) {
      if (!node || typeof node !== "object") {
        return Promise.resolve(null);
      }
      var toWalk = [node];
      var seen = [];
      var openPromises = [];
      var collectEntries = function() {
        var entries = [];
        for (var _i = 0; _i < arguments.length; _i++) {
          entries[_i] = arguments[_i];
        }
        for (var _a = 0, entries_1 = entries; _a < entries_1.length; _a++) {
          var entry = entries_1[_a];
          if (typeof entry === "object") {
            toWalk.push(entry);
          }
        }
      };
      var collectMapEntries = function() {
        var maps = [];
        for (var _i = 0; _i < arguments.length; _i++) {
          maps[_i] = arguments[_i];
        }
        for (var _a = 0, maps_1 = maps; _a < maps_1.length; _a++) {
          var map = maps_1[_a];
          if (typeof map === "object") {
            for (var k in map) {
              var key = k;
              var entry = map[key];
              if (typeof entry === "object") {
                toWalk.push(entry);
              }
            }
          }
        }
      };
      var collectArrayEntries = function() {
        var arrays = [];
        for (var _i = 0; _i < arguments.length; _i++) {
          arrays[_i] = arguments[_i];
        }
        for (var _a = 0, arrays_1 = arrays; _a < arrays_1.length; _a++) {
          var array = arrays_1[_a];
          if (Array.isArray(array)) {
            for (var _b = 0, array_1 = array; _b < array_1.length; _b++) {
              var entry = array_1[_b];
              if (typeof entry === "object") {
                toWalk.push(entry);
              }
            }
          }
        }
      };
      var handleRef = function(next2) {
        var seenRefs = [];
        while (next2.$ref) {
          var ref = next2.$ref;
          var segments = ref.split("#", 2);
          delete next2.$ref;
          if (segments[0].length > 0) {
            openPromises.push(resolveExternalLink(next2, segments[0], segments[1], parentSchemaURL, parentSchemaDependencies));
            return;
          } else {
            if (seenRefs.indexOf(ref) === -1) {
              merge(next2, parentSchema, parentSchemaURL, segments[1]);
              seenRefs.push(ref);
            }
          }
        }
        collectEntries(next2.items, next2.additionalItems, next2.additionalProperties, next2.not, next2.contains, next2.propertyNames, next2.if, next2.then, next2.else);
        collectMapEntries(next2.definitions, next2.properties, next2.patternProperties, next2.dependencies);
        collectArrayEntries(next2.anyOf, next2.allOf, next2.oneOf, next2.items);
      };
      while (toWalk.length) {
        var next = toWalk.pop();
        if (seen.indexOf(next) >= 0) {
          continue;
        }
        seen.push(next);
        handleRef(next);
      }
      return _this.promise.all(openPromises);
    };
    return resolveRefs(schema, schema, schemaURL, dependencies).then(function(_) {
      return new ResolvedSchema(schema, resolveErrors);
    });
  };
  JSONSchemaService2.prototype.getSchemaForResource = function(resource, document) {
    if (document && document.root && document.root.type === "object") {
      var schemaProperties = document.root.properties.filter(function(p) {
        return p.keyNode.value === "$schema" && p.valueNode && p.valueNode.type === "string";
      });
      if (schemaProperties.length > 0) {
        var valueNode = schemaProperties[0].valueNode;
        if (valueNode && valueNode.type === "string") {
          var schemeId = getNodeValue2(valueNode);
          if (schemeId && startsWith(schemeId, ".") && this.contextService) {
            schemeId = this.contextService.resolveRelativePath(schemeId, resource);
          }
          if (schemeId) {
            var id = normalizeId(schemeId);
            return this.getOrAddSchemaHandle(id).getResolvedSchema();
          }
        }
      }
    }
    if (this.cachedSchemaForResource && this.cachedSchemaForResource.resource === resource) {
      return this.cachedSchemaForResource.resolvedSchema;
    }
    var seen = Object.create(null);
    var schemas = [];
    var normalizedResource = normalizeResourceForMatching(resource);
    for (var _i = 0, _a = this.filePatternAssociations; _i < _a.length; _i++) {
      var entry = _a[_i];
      if (entry.matchesPattern(normalizedResource)) {
        for (var _b = 0, _c = entry.getURIs(); _b < _c.length; _b++) {
          var schemaId = _c[_b];
          if (!seen[schemaId]) {
            schemas.push(schemaId);
            seen[schemaId] = true;
          }
        }
      }
    }
    var resolvedSchema = schemas.length > 0 ? this.createCombinedSchema(resource, schemas).getResolvedSchema() : this.promise.resolve(void 0);
    this.cachedSchemaForResource = { resource, resolvedSchema };
    return resolvedSchema;
  };
  JSONSchemaService2.prototype.createCombinedSchema = function(resource, schemaIds) {
    if (schemaIds.length === 1) {
      return this.getOrAddSchemaHandle(schemaIds[0]);
    } else {
      var combinedSchemaId = "schemaservice://combinedSchema/" + encodeURIComponent(resource);
      var combinedSchema = {
        allOf: schemaIds.map(function(schemaId) {
          return { $ref: schemaId };
        })
      };
      return this.addSchemaHandle(combinedSchemaId, combinedSchema);
    }
  };
  JSONSchemaService2.prototype.getMatchingSchemas = function(document, jsonDocument, schema) {
    if (schema) {
      var id = schema.id || "schemaservice://untitled/matchingSchemas/" + idCounter++;
      return this.resolveSchemaContent(new UnresolvedSchema(schema), id, {}).then(function(resolvedSchema) {
        return jsonDocument.getMatchingSchemas(resolvedSchema.schema).filter(function(s) {
          return !s.inverted;
        });
      });
    }
    return this.getSchemaForResource(document.uri, jsonDocument).then(function(schema2) {
      if (schema2) {
        return jsonDocument.getMatchingSchemas(schema2.schema).filter(function(s) {
          return !s.inverted;
        });
      }
      return [];
    });
  };
  return JSONSchemaService2;
}();
var idCounter = 0;
function normalizeId(id) {
  try {
    return URI.parse(id).toString();
  } catch (e) {
    return id;
  }
}
function normalizeResourceForMatching(resource) {
  try {
    return URI.parse(resource).with({ fragment: null, query: null }).toString();
  } catch (e) {
    return resource;
  }
}
function toDisplayString(url) {
  try {
    var uri = URI.parse(url);
    if (uri.scheme === "file") {
      return uri.fsPath;
    }
  } catch (e) {
  }
  return url;
}

// node_modules/yaml-language-server/lib/esm/languageservice/utils/strings.js
"use strict";
function getIndentation(lineContent, position) {
  if (lineContent.length < position) {
    return 0;
  }
  for (let i = 0; i < position; i++) {
    const char = lineContent.charCodeAt(i);
    if (char !== 32 && char !== 9) {
      return i;
    }
  }
  return position;
}
function safeCreateUnicodeRegExp(pattern) {
  try {
    return new RegExp(pattern, "u");
  } catch (ignore) {
    return new RegExp(pattern);
  }
}

// node_modules/yaml-language-server/lib/esm/languageservice/services/yamlSchemaService.js
import { parse as parse4 } from "yaml";
import {
  isAbsolute,
  parse as parse5,
  resolve
} from "path-browserify";

// node_modules/yaml-language-server/lib/esm/languageservice/parser/yamlParser07.js
import { Parser as Parser5, Composer, LineCounter } from "yaml";

// node_modules/yaml-language-server/lib/esm/languageservice/parser/jsonParser07.js
import {
  findNodeAtOffset as findNodeAtOffset2,
  getNodePath as getNodePath3,
  getNodeValue as getNodeValue3
} from "jsonc-parser";

// node_modules/yaml-language-server/lib/esm/languageservice/utils/objects.js
"use strict";
function equals2(one, other) {
  if (one === other) {
    return true;
  }
  if (one === null || one === void 0 || other === null || other === void 0) {
    return false;
  }
  if (typeof one !== typeof other) {
    return false;
  }
  if (typeof one !== "object") {
    return false;
  }
  if (Array.isArray(one) !== Array.isArray(other)) {
    return false;
  }
  let i, key;
  if (Array.isArray(one)) {
    if (one.length !== other.length) {
      return false;
    }
    for (i = 0; i < one.length; i++) {
      if (!equals2(one[i], other[i])) {
        return false;
      }
    }
  } else {
    const oneKeys = [];
    for (key in one) {
      oneKeys.push(key);
    }
    oneKeys.sort();
    const otherKeys = [];
    for (key in other) {
      otherKeys.push(key);
    }
    otherKeys.sort();
    if (!equals2(oneKeys, otherKeys)) {
      return false;
    }
    for (i = 0; i < oneKeys.length; i++) {
      if (!equals2(one[oneKeys[i]], other[oneKeys[i]])) {
        return false;
      }
    }
  }
  return true;
}
function isNumber2(val) {
  return typeof val === "number";
}
function isDefined2(val) {
  return typeof val !== "undefined";
}
function isBoolean2(val) {
  return typeof val === "boolean";
}
function isString2(val) {
  return typeof val === "string";
}
function isIterable(val) {
  return Symbol.iterator in Object(val);
}

// node_modules/yaml-language-server/lib/esm/languageservice/utils/schemaUtils.js
function getSchemaTypeName(schema) {
  if (schema.$id) {
    const type = getSchemaRefTypeTitle(schema.$id);
    return type;
  }
  if (schema.$ref || schema._$ref) {
    const type = getSchemaRefTypeTitle(schema.$ref || schema._$ref);
    return type;
  }
  const typeStr = schema.title || (Array.isArray(schema.type) ? schema.type.join(" | ") : schema.type);
  return typeStr;
}
function getSchemaRefTypeTitle($ref) {
  const match = $ref.match(/^(?:.*\/)?(.*?)(?:\.schema\.json)?$/);
  let type = !!match && match[1];
  if (!type) {
    type = "typeNotFound";
    console.error(`$ref (${$ref}) not parsed properly`);
  }
  return type;
}

// node_modules/vscode-json-languageservice/lib/esm/services/jsonCompletion.js
import {
  createScanner as createScanner2
} from "jsonc-parser";

// node_modules/vscode-json-languageservice/lib/esm/utils/json.js
function stringifyObject(obj, indent, stringifyLiteral) {
  if (obj !== null && typeof obj === "object") {
    var newIndent = indent + "	";
    if (Array.isArray(obj)) {
      if (obj.length === 0) {
        return "[]";
      }
      var result = "[\n";
      for (var i = 0; i < obj.length; i++) {
        result += newIndent + stringifyObject(obj[i], newIndent, stringifyLiteral);
        if (i < obj.length - 1) {
          result += ",";
        }
        result += "\n";
      }
      result += indent + "]";
      return result;
    } else {
      var keys = Object.keys(obj);
      if (keys.length === 0) {
        return "{}";
      }
      var result = "{\n";
      for (var i = 0; i < keys.length; i++) {
        var key = keys[i];
        result += newIndent + JSON.stringify(key) + ": " + stringifyObject(obj[key], newIndent, stringifyLiteral);
        if (i < keys.length - 1) {
          result += ",";
        }
        result += "\n";
      }
      result += indent + "}";
      return result;
    }
  }
  return stringifyLiteral(obj);
}

// node_modules/vscode-json-languageservice/lib/esm/services/jsonCompletion.js
var localize4 = loadMessageBundle();
var valueCommitCharacters = [",", "}", "]"];
var propertyCommitCharacters = [":"];
var JSONCompletion = function() {
  function JSONCompletion2(schemaService, contributions, promiseConstructor, clientCapabilities) {
    if (contributions === void 0) {
      contributions = [];
    }
    if (promiseConstructor === void 0) {
      promiseConstructor = Promise;
    }
    if (clientCapabilities === void 0) {
      clientCapabilities = {};
    }
    this.schemaService = schemaService;
    this.contributions = contributions;
    this.promiseConstructor = promiseConstructor;
    this.clientCapabilities = clientCapabilities;
  }
  JSONCompletion2.prototype.doResolve = function(item) {
    for (var i = this.contributions.length - 1; i >= 0; i--) {
      var resolveCompletion = this.contributions[i].resolveCompletion;
      if (resolveCompletion) {
        var resolver = resolveCompletion(item);
        if (resolver) {
          return resolver;
        }
      }
    }
    return this.promiseConstructor.resolve(item);
  };
  JSONCompletion2.prototype.doComplete = function(document, position, doc) {
    var _this = this;
    var result = {
      items: [],
      isIncomplete: false
    };
    var text = document.getText();
    var offset = document.offsetAt(position);
    var node = doc.getNodeFromOffset(offset, true);
    if (this.isInComment(document, node ? node.offset : 0, offset)) {
      return Promise.resolve(result);
    }
    if (node && offset === node.offset + node.length && offset > 0) {
      var ch = text[offset - 1];
      if (node.type === "object" && ch === "}" || node.type === "array" && ch === "]") {
        node = node.parent;
      }
    }
    var currentWord = this.getCurrentWord(document, offset);
    var overwriteRange;
    if (node && (node.type === "string" || node.type === "number" || node.type === "boolean" || node.type === "null")) {
      overwriteRange = Range.create(document.positionAt(node.offset), document.positionAt(node.offset + node.length));
    } else {
      var overwriteStart = offset - currentWord.length;
      if (overwriteStart > 0 && text[overwriteStart - 1] === '"') {
        overwriteStart--;
      }
      overwriteRange = Range.create(document.positionAt(overwriteStart), position);
    }
    var supportsCommitCharacters = false;
    var proposed = {};
    var collector = {
      add: function(suggestion) {
        var label = suggestion.label;
        var existing = proposed[label];
        if (!existing) {
          label = label.replace(/[\n]/g, "\u21B5");
          if (label.length > 60) {
            var shortendedLabel = label.substr(0, 57).trim() + "...";
            if (!proposed[shortendedLabel]) {
              label = shortendedLabel;
            }
          }
          if (overwriteRange && suggestion.insertText !== void 0) {
            suggestion.textEdit = TextEdit.replace(overwriteRange, suggestion.insertText);
          }
          if (supportsCommitCharacters) {
            suggestion.commitCharacters = suggestion.kind === CompletionItemKind.Property ? propertyCommitCharacters : valueCommitCharacters;
          }
          suggestion.label = label;
          proposed[label] = suggestion;
          result.items.push(suggestion);
        } else {
          if (!existing.documentation) {
            existing.documentation = suggestion.documentation;
          }
          if (!existing.detail) {
            existing.detail = suggestion.detail;
          }
        }
      },
      setAsIncomplete: function() {
        result.isIncomplete = true;
      },
      error: function(message) {
        console.error(message);
      },
      log: function(message) {
        console.log(message);
      },
      getNumberOfProposals: function() {
        return result.items.length;
      }
    };
    return this.schemaService.getSchemaForResource(document.uri, doc).then(function(schema) {
      var collectionPromises = [];
      var addValue = true;
      var currentKey = "";
      var currentProperty = void 0;
      if (node) {
        if (node.type === "string") {
          var parent = node.parent;
          if (parent && parent.type === "property" && parent.keyNode === node) {
            addValue = !parent.valueNode;
            currentProperty = parent;
            currentKey = text.substr(node.offset + 1, node.length - 2);
            if (parent) {
              node = parent.parent;
            }
          }
        }
      }
      if (node && node.type === "object") {
        if (node.offset === offset) {
          return result;
        }
        var properties = node.properties;
        properties.forEach(function(p) {
          if (!currentProperty || currentProperty !== p) {
            proposed[p.keyNode.value] = CompletionItem.create("__");
          }
        });
        var separatorAfter_1 = "";
        if (addValue) {
          separatorAfter_1 = _this.evaluateSeparatorAfter(document, document.offsetAt(overwriteRange.end));
        }
        if (schema) {
          _this.getPropertyCompletions(schema, doc, node, addValue, separatorAfter_1, collector);
        } else {
          _this.getSchemaLessPropertyCompletions(doc, node, currentKey, collector);
        }
        var location_1 = getNodePath2(node);
        _this.contributions.forEach(function(contribution) {
          var collectPromise = contribution.collectPropertyCompletions(document.uri, location_1, currentWord, addValue, separatorAfter_1 === "", collector);
          if (collectPromise) {
            collectionPromises.push(collectPromise);
          }
        });
        if (!schema && currentWord.length > 0 && text.charAt(offset - currentWord.length - 1) !== '"') {
          collector.add({
            kind: CompletionItemKind.Property,
            label: _this.getLabelForValue(currentWord),
            insertText: _this.getInsertTextForProperty(currentWord, void 0, false, separatorAfter_1),
            insertTextFormat: InsertTextFormat.Snippet,
            documentation: ""
          });
          collector.setAsIncomplete();
        }
      }
      var types = {};
      if (schema) {
        _this.getValueCompletions(schema, doc, node, offset, document, collector, types);
      } else {
        _this.getSchemaLessValueCompletions(doc, node, offset, document, collector);
      }
      if (_this.contributions.length > 0) {
        _this.getContributedValueCompletions(doc, node, offset, document, collector, collectionPromises);
      }
      return _this.promiseConstructor.all(collectionPromises).then(function() {
        if (collector.getNumberOfProposals() === 0) {
          var offsetForSeparator = offset;
          if (node && (node.type === "string" || node.type === "number" || node.type === "boolean" || node.type === "null")) {
            offsetForSeparator = node.offset + node.length;
          }
          var separatorAfter = _this.evaluateSeparatorAfter(document, offsetForSeparator);
          _this.addFillerValueCompletions(types, separatorAfter, collector);
        }
        return result;
      });
    });
  };
  JSONCompletion2.prototype.getPropertyCompletions = function(schema, doc, node, addValue, separatorAfter, collector) {
    var _this = this;
    var matchingSchemas = doc.getMatchingSchemas(schema.schema, node.offset);
    matchingSchemas.forEach(function(s) {
      if (s.node === node && !s.inverted) {
        var schemaProperties_1 = s.schema.properties;
        if (schemaProperties_1) {
          Object.keys(schemaProperties_1).forEach(function(key) {
            var propertySchema = schemaProperties_1[key];
            if (typeof propertySchema === "object" && !propertySchema.deprecationMessage && !propertySchema.doNotSuggest) {
              var proposal = {
                kind: CompletionItemKind.Property,
                label: key,
                insertText: _this.getInsertTextForProperty(key, propertySchema, addValue, separatorAfter),
                insertTextFormat: InsertTextFormat.Snippet,
                filterText: _this.getFilterTextForValue(key),
                documentation: _this.fromMarkup(propertySchema.markdownDescription) || propertySchema.description || ""
              };
              if (propertySchema.suggestSortText !== void 0) {
                proposal.sortText = propertySchema.suggestSortText;
              }
              if (proposal.insertText && endsWith(proposal.insertText, "$1" + separatorAfter)) {
                proposal.command = {
                  title: "Suggest",
                  command: "editor.action.triggerSuggest"
                };
              }
              collector.add(proposal);
            }
          });
        }
        var schemaPropertyNames_1 = s.schema.propertyNames;
        if (typeof schemaPropertyNames_1 === "object" && !schemaPropertyNames_1.deprecationMessage && !schemaPropertyNames_1.doNotSuggest) {
          var propertyNameCompletionItem = function(name, enumDescription2) {
            if (enumDescription2 === void 0) {
              enumDescription2 = void 0;
            }
            var proposal = {
              kind: CompletionItemKind.Property,
              label: name,
              insertText: _this.getInsertTextForProperty(name, void 0, addValue, separatorAfter),
              insertTextFormat: InsertTextFormat.Snippet,
              filterText: _this.getFilterTextForValue(name),
              documentation: enumDescription2 || _this.fromMarkup(schemaPropertyNames_1.markdownDescription) || schemaPropertyNames_1.description || ""
            };
            if (schemaPropertyNames_1.suggestSortText !== void 0) {
              proposal.sortText = schemaPropertyNames_1.suggestSortText;
            }
            if (proposal.insertText && endsWith(proposal.insertText, "$1" + separatorAfter)) {
              proposal.command = {
                title: "Suggest",
                command: "editor.action.triggerSuggest"
              };
            }
            collector.add(proposal);
          };
          if (schemaPropertyNames_1.enum) {
            for (var i = 0; i < schemaPropertyNames_1.enum.length; i++) {
              var enumDescription = void 0;
              if (schemaPropertyNames_1.markdownEnumDescriptions && i < schemaPropertyNames_1.markdownEnumDescriptions.length) {
                enumDescription = _this.fromMarkup(schemaPropertyNames_1.markdownEnumDescriptions[i]);
              } else if (schemaPropertyNames_1.enumDescriptions && i < schemaPropertyNames_1.enumDescriptions.length) {
                enumDescription = schemaPropertyNames_1.enumDescriptions[i];
              }
              propertyNameCompletionItem(schemaPropertyNames_1.enum[i], enumDescription);
            }
          }
          if (schemaPropertyNames_1.const) {
            propertyNameCompletionItem(schemaPropertyNames_1.const);
          }
        }
      }
    });
  };
  JSONCompletion2.prototype.getSchemaLessPropertyCompletions = function(doc, node, currentKey, collector) {
    var _this = this;
    var collectCompletionsForSimilarObject = function(obj) {
      obj.properties.forEach(function(p) {
        var key = p.keyNode.value;
        collector.add({
          kind: CompletionItemKind.Property,
          label: key,
          insertText: _this.getInsertTextForValue(key, ""),
          insertTextFormat: InsertTextFormat.Snippet,
          filterText: _this.getFilterTextForValue(key),
          documentation: ""
        });
      });
    };
    if (node.parent) {
      if (node.parent.type === "property") {
        var parentKey_1 = node.parent.keyNode.value;
        doc.visit(function(n) {
          if (n.type === "property" && n !== node.parent && n.keyNode.value === parentKey_1 && n.valueNode && n.valueNode.type === "object") {
            collectCompletionsForSimilarObject(n.valueNode);
          }
          return true;
        });
      } else if (node.parent.type === "array") {
        node.parent.items.forEach(function(n) {
          if (n.type === "object" && n !== node) {
            collectCompletionsForSimilarObject(n);
          }
        });
      }
    } else if (node.type === "object") {
      collector.add({
        kind: CompletionItemKind.Property,
        label: "$schema",
        insertText: this.getInsertTextForProperty("$schema", void 0, true, ""),
        insertTextFormat: InsertTextFormat.Snippet,
        documentation: "",
        filterText: this.getFilterTextForValue("$schema")
      });
    }
  };
  JSONCompletion2.prototype.getSchemaLessValueCompletions = function(doc, node, offset, document, collector) {
    var _this = this;
    var offsetForSeparator = offset;
    if (node && (node.type === "string" || node.type === "number" || node.type === "boolean" || node.type === "null")) {
      offsetForSeparator = node.offset + node.length;
      node = node.parent;
    }
    if (!node) {
      collector.add({
        kind: this.getSuggestionKind("object"),
        label: "Empty object",
        insertText: this.getInsertTextForValue({}, ""),
        insertTextFormat: InsertTextFormat.Snippet,
        documentation: ""
      });
      collector.add({
        kind: this.getSuggestionKind("array"),
        label: "Empty array",
        insertText: this.getInsertTextForValue([], ""),
        insertTextFormat: InsertTextFormat.Snippet,
        documentation: ""
      });
      return;
    }
    var separatorAfter = this.evaluateSeparatorAfter(document, offsetForSeparator);
    var collectSuggestionsForValues = function(value) {
      if (value.parent && !contains(value.parent, offset, true)) {
        collector.add({
          kind: _this.getSuggestionKind(value.type),
          label: _this.getLabelTextForMatchingNode(value, document),
          insertText: _this.getInsertTextForMatchingNode(value, document, separatorAfter),
          insertTextFormat: InsertTextFormat.Snippet,
          documentation: ""
        });
      }
      if (value.type === "boolean") {
        _this.addBooleanValueCompletion(!value.value, separatorAfter, collector);
      }
    };
    if (node.type === "property") {
      if (offset > (node.colonOffset || 0)) {
        var valueNode = node.valueNode;
        if (valueNode && (offset > valueNode.offset + valueNode.length || valueNode.type === "object" || valueNode.type === "array")) {
          return;
        }
        var parentKey_2 = node.keyNode.value;
        doc.visit(function(n) {
          if (n.type === "property" && n.keyNode.value === parentKey_2 && n.valueNode) {
            collectSuggestionsForValues(n.valueNode);
          }
          return true;
        });
        if (parentKey_2 === "$schema" && node.parent && !node.parent.parent) {
          this.addDollarSchemaCompletions(separatorAfter, collector);
        }
      }
    }
    if (node.type === "array") {
      if (node.parent && node.parent.type === "property") {
        var parentKey_3 = node.parent.keyNode.value;
        doc.visit(function(n) {
          if (n.type === "property" && n.keyNode.value === parentKey_3 && n.valueNode && n.valueNode.type === "array") {
            n.valueNode.items.forEach(collectSuggestionsForValues);
          }
          return true;
        });
      } else {
        node.items.forEach(collectSuggestionsForValues);
      }
    }
  };
  JSONCompletion2.prototype.getValueCompletions = function(schema, doc, node, offset, document, collector, types) {
    var offsetForSeparator = offset;
    var parentKey = void 0;
    var valueNode = void 0;
    if (node && (node.type === "string" || node.type === "number" || node.type === "boolean" || node.type === "null")) {
      offsetForSeparator = node.offset + node.length;
      valueNode = node;
      node = node.parent;
    }
    if (!node) {
      this.addSchemaValueCompletions(schema.schema, "", collector, types);
      return;
    }
    if (node.type === "property" && offset > (node.colonOffset || 0)) {
      var valueNode_1 = node.valueNode;
      if (valueNode_1 && offset > valueNode_1.offset + valueNode_1.length) {
        return;
      }
      parentKey = node.keyNode.value;
      node = node.parent;
    }
    if (node && (parentKey !== void 0 || node.type === "array")) {
      var separatorAfter = this.evaluateSeparatorAfter(document, offsetForSeparator);
      var matchingSchemas = doc.getMatchingSchemas(schema.schema, node.offset, valueNode);
      for (var _i = 0, matchingSchemas_1 = matchingSchemas; _i < matchingSchemas_1.length; _i++) {
        var s = matchingSchemas_1[_i];
        if (s.node === node && !s.inverted && s.schema) {
          if (node.type === "array" && s.schema.items) {
            if (Array.isArray(s.schema.items)) {
              var index = this.findItemAtOffset(node, document, offset);
              if (index < s.schema.items.length) {
                this.addSchemaValueCompletions(s.schema.items[index], separatorAfter, collector, types);
              }
            } else {
              this.addSchemaValueCompletions(s.schema.items, separatorAfter, collector, types);
            }
          }
          if (parentKey !== void 0) {
            var propertyMatched = false;
            if (s.schema.properties) {
              var propertySchema = s.schema.properties[parentKey];
              if (propertySchema) {
                propertyMatched = true;
                this.addSchemaValueCompletions(propertySchema, separatorAfter, collector, types);
              }
            }
            if (s.schema.patternProperties && !propertyMatched) {
              for (var _a = 0, _b = Object.keys(s.schema.patternProperties); _a < _b.length; _a++) {
                var pattern = _b[_a];
                var regex = extendedRegExp(pattern);
                if (regex === null || regex === void 0 ? void 0 : regex.test(parentKey)) {
                  propertyMatched = true;
                  var propertySchema = s.schema.patternProperties[pattern];
                  this.addSchemaValueCompletions(propertySchema, separatorAfter, collector, types);
                }
              }
            }
            if (s.schema.additionalProperties && !propertyMatched) {
              var propertySchema = s.schema.additionalProperties;
              this.addSchemaValueCompletions(propertySchema, separatorAfter, collector, types);
            }
          }
        }
      }
      if (parentKey === "$schema" && !node.parent) {
        this.addDollarSchemaCompletions(separatorAfter, collector);
      }
      if (types["boolean"]) {
        this.addBooleanValueCompletion(true, separatorAfter, collector);
        this.addBooleanValueCompletion(false, separatorAfter, collector);
      }
      if (types["null"]) {
        this.addNullValueCompletion(separatorAfter, collector);
      }
    }
  };
  JSONCompletion2.prototype.getContributedValueCompletions = function(doc, node, offset, document, collector, collectionPromises) {
    if (!node) {
      this.contributions.forEach(function(contribution) {
        var collectPromise = contribution.collectDefaultCompletions(document.uri, collector);
        if (collectPromise) {
          collectionPromises.push(collectPromise);
        }
      });
    } else {
      if (node.type === "string" || node.type === "number" || node.type === "boolean" || node.type === "null") {
        node = node.parent;
      }
      if (node && node.type === "property" && offset > (node.colonOffset || 0)) {
        var parentKey_4 = node.keyNode.value;
        var valueNode = node.valueNode;
        if ((!valueNode || offset <= valueNode.offset + valueNode.length) && node.parent) {
          var location_2 = getNodePath2(node.parent);
          this.contributions.forEach(function(contribution) {
            var collectPromise = contribution.collectValueCompletions(document.uri, location_2, parentKey_4, collector);
            if (collectPromise) {
              collectionPromises.push(collectPromise);
            }
          });
        }
      }
    }
  };
  JSONCompletion2.prototype.addSchemaValueCompletions = function(schema, separatorAfter, collector, types) {
    var _this = this;
    if (typeof schema === "object") {
      this.addEnumValueCompletions(schema, separatorAfter, collector);
      this.addDefaultValueCompletions(schema, separatorAfter, collector);
      this.collectTypes(schema, types);
      if (Array.isArray(schema.allOf)) {
        schema.allOf.forEach(function(s) {
          return _this.addSchemaValueCompletions(s, separatorAfter, collector, types);
        });
      }
      if (Array.isArray(schema.anyOf)) {
        schema.anyOf.forEach(function(s) {
          return _this.addSchemaValueCompletions(s, separatorAfter, collector, types);
        });
      }
      if (Array.isArray(schema.oneOf)) {
        schema.oneOf.forEach(function(s) {
          return _this.addSchemaValueCompletions(s, separatorAfter, collector, types);
        });
      }
    }
  };
  JSONCompletion2.prototype.addDefaultValueCompletions = function(schema, separatorAfter, collector, arrayDepth) {
    var _this = this;
    if (arrayDepth === void 0) {
      arrayDepth = 0;
    }
    var hasProposals = false;
    if (isDefined(schema.default)) {
      var type = schema.type;
      var value = schema.default;
      for (var i = arrayDepth; i > 0; i--) {
        value = [value];
        type = "array";
      }
      collector.add({
        kind: this.getSuggestionKind(type),
        label: this.getLabelForValue(value),
        insertText: this.getInsertTextForValue(value, separatorAfter),
        insertTextFormat: InsertTextFormat.Snippet,
        detail: localize4("json.suggest.default", "Default value")
      });
      hasProposals = true;
    }
    if (Array.isArray(schema.examples)) {
      schema.examples.forEach(function(example) {
        var type2 = schema.type;
        var value2 = example;
        for (var i2 = arrayDepth; i2 > 0; i2--) {
          value2 = [value2];
          type2 = "array";
        }
        collector.add({
          kind: _this.getSuggestionKind(type2),
          label: _this.getLabelForValue(value2),
          insertText: _this.getInsertTextForValue(value2, separatorAfter),
          insertTextFormat: InsertTextFormat.Snippet
        });
        hasProposals = true;
      });
    }
    if (Array.isArray(schema.defaultSnippets)) {
      schema.defaultSnippets.forEach(function(s) {
        var type2 = schema.type;
        var value2 = s.body;
        var label = s.label;
        var insertText;
        var filterText;
        if (isDefined(value2)) {
          var type_1 = schema.type;
          for (var i2 = arrayDepth; i2 > 0; i2--) {
            value2 = [value2];
            type_1 = "array";
          }
          insertText = _this.getInsertTextForSnippetValue(value2, separatorAfter);
          filterText = _this.getFilterTextForSnippetValue(value2);
          label = label || _this.getLabelForSnippetValue(value2);
        } else if (typeof s.bodyText === "string") {
          var prefix = "", suffix = "", indent = "";
          for (var i2 = arrayDepth; i2 > 0; i2--) {
            prefix = prefix + indent + "[\n";
            suffix = suffix + "\n" + indent + "]";
            indent += "	";
            type2 = "array";
          }
          insertText = prefix + indent + s.bodyText.split("\n").join("\n" + indent) + suffix + separatorAfter;
          label = label || insertText, filterText = insertText.replace(/[\n]/g, "");
        } else {
          return;
        }
        collector.add({
          kind: _this.getSuggestionKind(type2),
          label,
          documentation: _this.fromMarkup(s.markdownDescription) || s.description,
          insertText,
          insertTextFormat: InsertTextFormat.Snippet,
          filterText
        });
        hasProposals = true;
      });
    }
    if (!hasProposals && typeof schema.items === "object" && !Array.isArray(schema.items) && arrayDepth < 5) {
      this.addDefaultValueCompletions(schema.items, separatorAfter, collector, arrayDepth + 1);
    }
  };
  JSONCompletion2.prototype.addEnumValueCompletions = function(schema, separatorAfter, collector) {
    if (isDefined(schema.const)) {
      collector.add({
        kind: this.getSuggestionKind(schema.type),
        label: this.getLabelForValue(schema.const),
        insertText: this.getInsertTextForValue(schema.const, separatorAfter),
        insertTextFormat: InsertTextFormat.Snippet,
        documentation: this.fromMarkup(schema.markdownDescription) || schema.description
      });
    }
    if (Array.isArray(schema.enum)) {
      for (var i = 0, length = schema.enum.length; i < length; i++) {
        var enm = schema.enum[i];
        var documentation = this.fromMarkup(schema.markdownDescription) || schema.description;
        if (schema.markdownEnumDescriptions && i < schema.markdownEnumDescriptions.length && this.doesSupportMarkdown()) {
          documentation = this.fromMarkup(schema.markdownEnumDescriptions[i]);
        } else if (schema.enumDescriptions && i < schema.enumDescriptions.length) {
          documentation = schema.enumDescriptions[i];
        }
        collector.add({
          kind: this.getSuggestionKind(schema.type),
          label: this.getLabelForValue(enm),
          insertText: this.getInsertTextForValue(enm, separatorAfter),
          insertTextFormat: InsertTextFormat.Snippet,
          documentation
        });
      }
    }
  };
  JSONCompletion2.prototype.collectTypes = function(schema, types) {
    if (Array.isArray(schema.enum) || isDefined(schema.const)) {
      return;
    }
    var type = schema.type;
    if (Array.isArray(type)) {
      type.forEach(function(t) {
        return types[t] = true;
      });
    } else if (type) {
      types[type] = true;
    }
  };
  JSONCompletion2.prototype.addFillerValueCompletions = function(types, separatorAfter, collector) {
    if (types["object"]) {
      collector.add({
        kind: this.getSuggestionKind("object"),
        label: "{}",
        insertText: this.getInsertTextForGuessedValue({}, separatorAfter),
        insertTextFormat: InsertTextFormat.Snippet,
        detail: localize4("defaults.object", "New object"),
        documentation: ""
      });
    }
    if (types["array"]) {
      collector.add({
        kind: this.getSuggestionKind("array"),
        label: "[]",
        insertText: this.getInsertTextForGuessedValue([], separatorAfter),
        insertTextFormat: InsertTextFormat.Snippet,
        detail: localize4("defaults.array", "New array"),
        documentation: ""
      });
    }
  };
  JSONCompletion2.prototype.addBooleanValueCompletion = function(value, separatorAfter, collector) {
    collector.add({
      kind: this.getSuggestionKind("boolean"),
      label: value ? "true" : "false",
      insertText: this.getInsertTextForValue(value, separatorAfter),
      insertTextFormat: InsertTextFormat.Snippet,
      documentation: ""
    });
  };
  JSONCompletion2.prototype.addNullValueCompletion = function(separatorAfter, collector) {
    collector.add({
      kind: this.getSuggestionKind("null"),
      label: "null",
      insertText: "null" + separatorAfter,
      insertTextFormat: InsertTextFormat.Snippet,
      documentation: ""
    });
  };
  JSONCompletion2.prototype.addDollarSchemaCompletions = function(separatorAfter, collector) {
    var _this = this;
    var schemaIds = this.schemaService.getRegisteredSchemaIds(function(schema) {
      return schema === "http" || schema === "https";
    });
    schemaIds.forEach(function(schemaId) {
      return collector.add({
        kind: CompletionItemKind.Module,
        label: _this.getLabelForValue(schemaId),
        filterText: _this.getFilterTextForValue(schemaId),
        insertText: _this.getInsertTextForValue(schemaId, separatorAfter),
        insertTextFormat: InsertTextFormat.Snippet,
        documentation: ""
      });
    });
  };
  JSONCompletion2.prototype.getLabelForValue = function(value) {
    return JSON.stringify(value);
  };
  JSONCompletion2.prototype.getFilterTextForValue = function(value) {
    return JSON.stringify(value);
  };
  JSONCompletion2.prototype.getFilterTextForSnippetValue = function(value) {
    return JSON.stringify(value).replace(/\$\{\d+:([^}]+)\}|\$\d+/g, "$1");
  };
  JSONCompletion2.prototype.getLabelForSnippetValue = function(value) {
    var label = JSON.stringify(value);
    return label.replace(/\$\{\d+:([^}]+)\}|\$\d+/g, "$1");
  };
  JSONCompletion2.prototype.getInsertTextForPlainText = function(text) {
    return text.replace(/[\\\$\}]/g, "\\$&");
  };
  JSONCompletion2.prototype.getInsertTextForValue = function(value, separatorAfter) {
    var text = JSON.stringify(value, null, "	");
    if (text === "{}") {
      return "{$1}" + separatorAfter;
    } else if (text === "[]") {
      return "[$1]" + separatorAfter;
    }
    return this.getInsertTextForPlainText(text + separatorAfter);
  };
  JSONCompletion2.prototype.getInsertTextForSnippetValue = function(value, separatorAfter) {
    var replacer = function(value2) {
      if (typeof value2 === "string") {
        if (value2[0] === "^") {
          return value2.substr(1);
        }
      }
      return JSON.stringify(value2);
    };
    return stringifyObject(value, "", replacer) + separatorAfter;
  };
  JSONCompletion2.prototype.getInsertTextForGuessedValue = function(value, separatorAfter) {
    switch (typeof value) {
      case "object":
        if (value === null) {
          return "${1:null}" + separatorAfter;
        }
        return this.getInsertTextForValue(value, separatorAfter);
      case "string":
        var snippetValue = JSON.stringify(value);
        snippetValue = snippetValue.substr(1, snippetValue.length - 2);
        snippetValue = this.getInsertTextForPlainText(snippetValue);
        return '"${1:' + snippetValue + '}"' + separatorAfter;
      case "number":
      case "boolean":
        return "${1:" + JSON.stringify(value) + "}" + separatorAfter;
    }
    return this.getInsertTextForValue(value, separatorAfter);
  };
  JSONCompletion2.prototype.getSuggestionKind = function(type) {
    if (Array.isArray(type)) {
      var array = type;
      type = array.length > 0 ? array[0] : void 0;
    }
    if (!type) {
      return CompletionItemKind.Value;
    }
    switch (type) {
      case "string":
        return CompletionItemKind.Value;
      case "object":
        return CompletionItemKind.Module;
      case "property":
        return CompletionItemKind.Property;
      default:
        return CompletionItemKind.Value;
    }
  };
  JSONCompletion2.prototype.getLabelTextForMatchingNode = function(node, document) {
    switch (node.type) {
      case "array":
        return "[]";
      case "object":
        return "{}";
      default:
        var content = document.getText().substr(node.offset, node.length);
        return content;
    }
  };
  JSONCompletion2.prototype.getInsertTextForMatchingNode = function(node, document, separatorAfter) {
    switch (node.type) {
      case "array":
        return this.getInsertTextForValue([], separatorAfter);
      case "object":
        return this.getInsertTextForValue({}, separatorAfter);
      default:
        var content = document.getText().substr(node.offset, node.length) + separatorAfter;
        return this.getInsertTextForPlainText(content);
    }
  };
  JSONCompletion2.prototype.getInsertTextForProperty = function(key, propertySchema, addValue, separatorAfter) {
    var propertyText = this.getInsertTextForValue(key, "");
    if (!addValue) {
      return propertyText;
    }
    var resultText = propertyText + ": ";
    var value;
    var nValueProposals = 0;
    if (propertySchema) {
      if (Array.isArray(propertySchema.defaultSnippets)) {
        if (propertySchema.defaultSnippets.length === 1) {
          var body = propertySchema.defaultSnippets[0].body;
          if (isDefined(body)) {
            value = this.getInsertTextForSnippetValue(body, "");
          }
        }
        nValueProposals += propertySchema.defaultSnippets.length;
      }
      if (propertySchema.enum) {
        if (!value && propertySchema.enum.length === 1) {
          value = this.getInsertTextForGuessedValue(propertySchema.enum[0], "");
        }
        nValueProposals += propertySchema.enum.length;
      }
      if (isDefined(propertySchema.default)) {
        if (!value) {
          value = this.getInsertTextForGuessedValue(propertySchema.default, "");
        }
        nValueProposals++;
      }
      if (Array.isArray(propertySchema.examples) && propertySchema.examples.length) {
        if (!value) {
          value = this.getInsertTextForGuessedValue(propertySchema.examples[0], "");
        }
        nValueProposals += propertySchema.examples.length;
      }
      if (nValueProposals === 0) {
        var type = Array.isArray(propertySchema.type) ? propertySchema.type[0] : propertySchema.type;
        if (!type) {
          if (propertySchema.properties) {
            type = "object";
          } else if (propertySchema.items) {
            type = "array";
          }
        }
        switch (type) {
          case "boolean":
            value = "$1";
            break;
          case "string":
            value = '"$1"';
            break;
          case "object":
            value = "{$1}";
            break;
          case "array":
            value = "[$1]";
            break;
          case "number":
          case "integer":
            value = "${1:0}";
            break;
          case "null":
            value = "${1:null}";
            break;
          default:
            return propertyText;
        }
      }
    }
    if (!value || nValueProposals > 1) {
      value = "$1";
    }
    return resultText + value + separatorAfter;
  };
  JSONCompletion2.prototype.getCurrentWord = function(document, offset) {
    var i = offset - 1;
    var text = document.getText();
    while (i >= 0 && ' 	\n\r\v":{[,]}'.indexOf(text.charAt(i)) === -1) {
      i--;
    }
    return text.substring(i + 1, offset);
  };
  JSONCompletion2.prototype.evaluateSeparatorAfter = function(document, offset) {
    var scanner = createScanner2(document.getText(), true);
    scanner.setPosition(offset);
    var token = scanner.scan();
    switch (token) {
      case 5:
      case 2:
      case 4:
      case 17:
        return "";
      default:
        return ",";
    }
  };
  JSONCompletion2.prototype.findItemAtOffset = function(node, document, offset) {
    var scanner = createScanner2(document.getText(), true);
    var children = node.items;
    for (var i = children.length - 1; i >= 0; i--) {
      var child = children[i];
      if (offset > child.offset + child.length) {
        scanner.setPosition(child.offset + child.length);
        var token = scanner.scan();
        if (token === 5 && offset >= scanner.getTokenOffset() + scanner.getTokenLength()) {
          return i + 1;
        }
        return i;
      } else if (offset >= child.offset) {
        return i;
      }
    }
    return 0;
  };
  JSONCompletion2.prototype.isInComment = function(document, start, offset) {
    var scanner = createScanner2(document.getText(), false);
    scanner.setPosition(start);
    var token = scanner.scan();
    while (token !== 17 && scanner.getTokenOffset() + scanner.getTokenLength() < offset) {
      token = scanner.scan();
    }
    return (token === 12 || token === 13) && scanner.getTokenOffset() <= offset;
  };
  JSONCompletion2.prototype.fromMarkup = function(markupString) {
    if (markupString && this.doesSupportMarkdown()) {
      return {
        kind: MarkupKind.Markdown,
        value: markupString
      };
    }
    return void 0;
  };
  JSONCompletion2.prototype.doesSupportMarkdown = function() {
    if (!isDefined(this.supportsMarkdown)) {
      var completion = this.clientCapabilities.textDocument && this.clientCapabilities.textDocument.completion;
      this.supportsMarkdown = completion && completion.completionItem && Array.isArray(completion.completionItem.documentationFormat) && completion.completionItem.documentationFormat.indexOf(MarkupKind.Markdown) !== -1;
    }
    return this.supportsMarkdown;
  };
  JSONCompletion2.prototype.doesSupportsCommitCharacters = function() {
    if (!isDefined(this.supportsCommitCharacters)) {
      var completion = this.clientCapabilities.textDocument && this.clientCapabilities.textDocument.completion;
      this.supportsCommitCharacters = completion && completion.completionItem && !!completion.completionItem.commitCharactersSupport;
    }
    return this.supportsCommitCharacters;
  };
  return JSONCompletion2;
}();

// node_modules/vscode-json-languageservice/lib/esm/services/jsonHover.js
var JSONHover = function() {
  function JSONHover2(schemaService, contributions, promiseConstructor) {
    if (contributions === void 0) {
      contributions = [];
    }
    this.schemaService = schemaService;
    this.contributions = contributions;
    this.promise = promiseConstructor || Promise;
  }
  JSONHover2.prototype.doHover = function(document, position, doc) {
    var offset = document.offsetAt(position);
    var node = doc.getNodeFromOffset(offset);
    if (!node || (node.type === "object" || node.type === "array") && offset > node.offset + 1 && offset < node.offset + node.length - 1) {
      return this.promise.resolve(null);
    }
    var hoverRangeNode = node;
    if (node.type === "string") {
      var parent = node.parent;
      if (parent && parent.type === "property" && parent.keyNode === node) {
        node = parent.valueNode;
        if (!node) {
          return this.promise.resolve(null);
        }
      }
    }
    var hoverRange = Range.create(document.positionAt(hoverRangeNode.offset), document.positionAt(hoverRangeNode.offset + hoverRangeNode.length));
    var createHover = function(contents) {
      var result = {
        contents,
        range: hoverRange
      };
      return result;
    };
    var location = getNodePath2(node);
    for (var i = this.contributions.length - 1; i >= 0; i--) {
      var contribution = this.contributions[i];
      var promise = contribution.getInfoContribution(document.uri, location);
      if (promise) {
        return promise.then(function(htmlContent) {
          return createHover(htmlContent);
        });
      }
    }
    return this.schemaService.getSchemaForResource(document.uri, doc).then(function(schema) {
      if (schema && node) {
        var matchingSchemas = doc.getMatchingSchemas(schema.schema, node.offset);
        var title_1 = void 0;
        var markdownDescription_1 = void 0;
        var markdownEnumValueDescription_1 = void 0, enumValue_1 = void 0;
        matchingSchemas.every(function(s) {
          if (s.node === node && !s.inverted && s.schema) {
            title_1 = title_1 || s.schema.title;
            markdownDescription_1 = markdownDescription_1 || s.schema.markdownDescription || toMarkdown(s.schema.description);
            if (s.schema.enum) {
              var idx = s.schema.enum.indexOf(getNodeValue2(node));
              if (s.schema.markdownEnumDescriptions) {
                markdownEnumValueDescription_1 = s.schema.markdownEnumDescriptions[idx];
              } else if (s.schema.enumDescriptions) {
                markdownEnumValueDescription_1 = toMarkdown(s.schema.enumDescriptions[idx]);
              }
              if (markdownEnumValueDescription_1) {
                enumValue_1 = s.schema.enum[idx];
                if (typeof enumValue_1 !== "string") {
                  enumValue_1 = JSON.stringify(enumValue_1);
                }
              }
            }
          }
          return true;
        });
        var result = "";
        if (title_1) {
          result = toMarkdown(title_1);
        }
        if (markdownDescription_1) {
          if (result.length > 0) {
            result += "\n\n";
          }
          result += markdownDescription_1;
        }
        if (markdownEnumValueDescription_1) {
          if (result.length > 0) {
            result += "\n\n";
          }
          result += "`" + toMarkdownCodeBlock(enumValue_1) + "`: " + markdownEnumValueDescription_1;
        }
        return createHover([result]);
      }
      return null;
    });
  };
  return JSONHover2;
}();
function toMarkdown(plain) {
  if (plain) {
    var res = plain.replace(/([^\n\r])(\r?\n)([^\n\r])/gm, "$1\n\n$3");
    return res.replace(/[\\`*_{}[\]()#+\-.!]/g, "\\$&");
  }
  return void 0;
}
function toMarkdownCodeBlock(content) {
  if (content.indexOf("`") !== -1) {
    return "`` " + content + " ``";
  }
  return content;
}

// node_modules/vscode-json-languageservice/lib/esm/services/jsonValidation.js
var localize5 = loadMessageBundle();
var JSONValidation = function() {
  function JSONValidation2(jsonSchemaService, promiseConstructor) {
    this.jsonSchemaService = jsonSchemaService;
    this.promise = promiseConstructor;
    this.validationEnabled = true;
  }
  JSONValidation2.prototype.configure = function(raw) {
    if (raw) {
      this.validationEnabled = raw.validate !== false;
      this.commentSeverity = raw.allowComments ? void 0 : DiagnosticSeverity.Error;
    }
  };
  JSONValidation2.prototype.doValidation = function(textDocument, jsonDocument, documentSettings, schema) {
    var _this = this;
    if (!this.validationEnabled) {
      return this.promise.resolve([]);
    }
    var diagnostics = [];
    var added = {};
    var addProblem = function(problem) {
      var signature = problem.range.start.line + " " + problem.range.start.character + " " + problem.message;
      if (!added[signature]) {
        added[signature] = true;
        diagnostics.push(problem);
      }
    };
    var getDiagnostics = function(schema2) {
      var trailingCommaSeverity = (documentSettings === null || documentSettings === void 0 ? void 0 : documentSettings.trailingCommas) ? toDiagnosticSeverity(documentSettings.trailingCommas) : DiagnosticSeverity.Error;
      var commentSeverity = (documentSettings === null || documentSettings === void 0 ? void 0 : documentSettings.comments) ? toDiagnosticSeverity(documentSettings.comments) : _this.commentSeverity;
      var schemaValidation = (documentSettings === null || documentSettings === void 0 ? void 0 : documentSettings.schemaValidation) ? toDiagnosticSeverity(documentSettings.schemaValidation) : DiagnosticSeverity.Warning;
      var schemaRequest = (documentSettings === null || documentSettings === void 0 ? void 0 : documentSettings.schemaRequest) ? toDiagnosticSeverity(documentSettings.schemaRequest) : DiagnosticSeverity.Warning;
      if (schema2) {
        if (schema2.errors.length && jsonDocument.root && schemaRequest) {
          var astRoot = jsonDocument.root;
          var property = astRoot.type === "object" ? astRoot.properties[0] : void 0;
          if (property && property.keyNode.value === "$schema") {
            var node = property.valueNode || property;
            var range = Range.create(textDocument.positionAt(node.offset), textDocument.positionAt(node.offset + node.length));
            addProblem(Diagnostic.create(range, schema2.errors[0], schemaRequest, ErrorCode.SchemaResolveError));
          } else {
            var range = Range.create(textDocument.positionAt(astRoot.offset), textDocument.positionAt(astRoot.offset + 1));
            addProblem(Diagnostic.create(range, schema2.errors[0], schemaRequest, ErrorCode.SchemaResolveError));
          }
        } else if (schemaValidation) {
          var semanticErrors = jsonDocument.validate(textDocument, schema2.schema, schemaValidation);
          if (semanticErrors) {
            semanticErrors.forEach(addProblem);
          }
        }
        if (schemaAllowsComments(schema2.schema)) {
          commentSeverity = void 0;
        }
        if (schemaAllowsTrailingCommas(schema2.schema)) {
          trailingCommaSeverity = void 0;
        }
      }
      for (var _i = 0, _a = jsonDocument.syntaxErrors; _i < _a.length; _i++) {
        var p = _a[_i];
        if (p.code === ErrorCode.TrailingComma) {
          if (typeof trailingCommaSeverity !== "number") {
            continue;
          }
          p.severity = trailingCommaSeverity;
        }
        addProblem(p);
      }
      if (typeof commentSeverity === "number") {
        var message_1 = localize5("InvalidCommentToken", "Comments are not permitted in JSON.");
        jsonDocument.comments.forEach(function(c) {
          addProblem(Diagnostic.create(c, message_1, commentSeverity, ErrorCode.CommentNotPermitted));
        });
      }
      return diagnostics;
    };
    if (schema) {
      var id = schema.id || "schemaservice://untitled/" + idCounter2++;
      return this.jsonSchemaService.resolveSchemaContent(new UnresolvedSchema(schema), id, {}).then(function(resolvedSchema) {
        return getDiagnostics(resolvedSchema);
      });
    }
    return this.jsonSchemaService.getSchemaForResource(textDocument.uri, jsonDocument).then(function(schema2) {
      return getDiagnostics(schema2);
    });
  };
  return JSONValidation2;
}();
var idCounter2 = 0;
function schemaAllowsComments(schemaRef) {
  if (schemaRef && typeof schemaRef === "object") {
    if (isBoolean(schemaRef.allowComments)) {
      return schemaRef.allowComments;
    }
    if (schemaRef.allOf) {
      for (var _i = 0, _a = schemaRef.allOf; _i < _a.length; _i++) {
        var schema = _a[_i];
        var allow = schemaAllowsComments(schema);
        if (isBoolean(allow)) {
          return allow;
        }
      }
    }
  }
  return void 0;
}
function schemaAllowsTrailingCommas(schemaRef) {
  if (schemaRef && typeof schemaRef === "object") {
    if (isBoolean(schemaRef.allowTrailingCommas)) {
      return schemaRef.allowTrailingCommas;
    }
    var deprSchemaRef = schemaRef;
    if (isBoolean(deprSchemaRef["allowsTrailingCommas"])) {
      return deprSchemaRef["allowsTrailingCommas"];
    }
    if (schemaRef.allOf) {
      for (var _i = 0, _a = schemaRef.allOf; _i < _a.length; _i++) {
        var schema = _a[_i];
        var allow = schemaAllowsTrailingCommas(schema);
        if (isBoolean(allow)) {
          return allow;
        }
      }
    }
  }
  return void 0;
}
function toDiagnosticSeverity(severityLevel) {
  switch (severityLevel) {
    case "error":
      return DiagnosticSeverity.Error;
    case "warning":
      return DiagnosticSeverity.Warning;
    case "ignore":
      return void 0;
  }
  return void 0;
}

// node_modules/vscode-json-languageservice/lib/esm/utils/colors.js
var Digit0 = 48;
var Digit9 = 57;
var A = 65;
var a = 97;
var f = 102;
function hexDigit(charCode) {
  if (charCode < Digit0) {
    return 0;
  }
  if (charCode <= Digit9) {
    return charCode - Digit0;
  }
  if (charCode < a) {
    charCode += a - A;
  }
  if (charCode >= a && charCode <= f) {
    return charCode - a + 10;
  }
  return 0;
}
function colorFromHex(text) {
  if (text[0] !== "#") {
    return void 0;
  }
  switch (text.length) {
    case 4:
      return {
        red: hexDigit(text.charCodeAt(1)) * 17 / 255,
        green: hexDigit(text.charCodeAt(2)) * 17 / 255,
        blue: hexDigit(text.charCodeAt(3)) * 17 / 255,
        alpha: 1
      };
    case 5:
      return {
        red: hexDigit(text.charCodeAt(1)) * 17 / 255,
        green: hexDigit(text.charCodeAt(2)) * 17 / 255,
        blue: hexDigit(text.charCodeAt(3)) * 17 / 255,
        alpha: hexDigit(text.charCodeAt(4)) * 17 / 255
      };
    case 7:
      return {
        red: (hexDigit(text.charCodeAt(1)) * 16 + hexDigit(text.charCodeAt(2))) / 255,
        green: (hexDigit(text.charCodeAt(3)) * 16 + hexDigit(text.charCodeAt(4))) / 255,
        blue: (hexDigit(text.charCodeAt(5)) * 16 + hexDigit(text.charCodeAt(6))) / 255,
        alpha: 1
      };
    case 9:
      return {
        red: (hexDigit(text.charCodeAt(1)) * 16 + hexDigit(text.charCodeAt(2))) / 255,
        green: (hexDigit(text.charCodeAt(3)) * 16 + hexDigit(text.charCodeAt(4))) / 255,
        blue: (hexDigit(text.charCodeAt(5)) * 16 + hexDigit(text.charCodeAt(6))) / 255,
        alpha: (hexDigit(text.charCodeAt(7)) * 16 + hexDigit(text.charCodeAt(8))) / 255
      };
  }
  return void 0;
}

// node_modules/vscode-json-languageservice/lib/esm/services/jsonDocumentSymbols.js
var JSONDocumentSymbols = function() {
  function JSONDocumentSymbols2(schemaService) {
    this.schemaService = schemaService;
  }
  JSONDocumentSymbols2.prototype.findDocumentSymbols = function(document, doc, context) {
    var _this = this;
    if (context === void 0) {
      context = { resultLimit: Number.MAX_VALUE };
    }
    var root = doc.root;
    if (!root) {
      return [];
    }
    var limit = context.resultLimit || Number.MAX_VALUE;
    var resourceString = document.uri;
    if (resourceString === "vscode://defaultsettings/keybindings.json" || endsWith(resourceString.toLowerCase(), "/user/keybindings.json")) {
      if (root.type === "array") {
        var result_1 = [];
        for (var _i = 0, _a = root.items; _i < _a.length; _i++) {
          var item = _a[_i];
          if (item.type === "object") {
            for (var _b = 0, _c = item.properties; _b < _c.length; _b++) {
              var property = _c[_b];
              if (property.keyNode.value === "key" && property.valueNode) {
                var location = Location.create(document.uri, getRange(document, item));
                result_1.push({ name: getNodeValue2(property.valueNode), kind: SymbolKind.Function, location });
                limit--;
                if (limit <= 0) {
                  if (context && context.onResultLimitExceeded) {
                    context.onResultLimitExceeded(resourceString);
                  }
                  return result_1;
                }
              }
            }
          }
        }
        return result_1;
      }
    }
    var toVisit = [
      { node: root, containerName: "" }
    ];
    var nextToVisit = 0;
    var limitExceeded = false;
    var result = [];
    var collectOutlineEntries = function(node, containerName) {
      if (node.type === "array") {
        node.items.forEach(function(node2) {
          if (node2) {
            toVisit.push({ node: node2, containerName });
          }
        });
      } else if (node.type === "object") {
        node.properties.forEach(function(property2) {
          var valueNode = property2.valueNode;
          if (valueNode) {
            if (limit > 0) {
              limit--;
              var location2 = Location.create(document.uri, getRange(document, property2));
              var childContainerName = containerName ? containerName + "." + property2.keyNode.value : property2.keyNode.value;
              result.push({ name: _this.getKeyLabel(property2), kind: _this.getSymbolKind(valueNode.type), location: location2, containerName });
              toVisit.push({ node: valueNode, containerName: childContainerName });
            } else {
              limitExceeded = true;
            }
          }
        });
      }
    };
    while (nextToVisit < toVisit.length) {
      var next = toVisit[nextToVisit++];
      collectOutlineEntries(next.node, next.containerName);
    }
    if (limitExceeded && context && context.onResultLimitExceeded) {
      context.onResultLimitExceeded(resourceString);
    }
    return result;
  };
  JSONDocumentSymbols2.prototype.findDocumentSymbols2 = function(document, doc, context) {
    var _this = this;
    if (context === void 0) {
      context = { resultLimit: Number.MAX_VALUE };
    }
    var root = doc.root;
    if (!root) {
      return [];
    }
    var limit = context.resultLimit || Number.MAX_VALUE;
    var resourceString = document.uri;
    if (resourceString === "vscode://defaultsettings/keybindings.json" || endsWith(resourceString.toLowerCase(), "/user/keybindings.json")) {
      if (root.type === "array") {
        var result_2 = [];
        for (var _i = 0, _a = root.items; _i < _a.length; _i++) {
          var item = _a[_i];
          if (item.type === "object") {
            for (var _b = 0, _c = item.properties; _b < _c.length; _b++) {
              var property = _c[_b];
              if (property.keyNode.value === "key" && property.valueNode) {
                var range = getRange(document, item);
                var selectionRange = getRange(document, property.keyNode);
                result_2.push({ name: getNodeValue2(property.valueNode), kind: SymbolKind.Function, range, selectionRange });
                limit--;
                if (limit <= 0) {
                  if (context && context.onResultLimitExceeded) {
                    context.onResultLimitExceeded(resourceString);
                  }
                  return result_2;
                }
              }
            }
          }
        }
        return result_2;
      }
    }
    var result = [];
    var toVisit = [
      { node: root, result }
    ];
    var nextToVisit = 0;
    var limitExceeded = false;
    var collectOutlineEntries = function(node, result2) {
      if (node.type === "array") {
        node.items.forEach(function(node2, index) {
          if (node2) {
            if (limit > 0) {
              limit--;
              var range2 = getRange(document, node2);
              var selectionRange2 = range2;
              var name = String(index);
              var symbol = { name, kind: _this.getSymbolKind(node2.type), range: range2, selectionRange: selectionRange2, children: [] };
              result2.push(symbol);
              toVisit.push({ result: symbol.children, node: node2 });
            } else {
              limitExceeded = true;
            }
          }
        });
      } else if (node.type === "object") {
        node.properties.forEach(function(property2) {
          var valueNode = property2.valueNode;
          if (valueNode) {
            if (limit > 0) {
              limit--;
              var range2 = getRange(document, property2);
              var selectionRange2 = getRange(document, property2.keyNode);
              var children = [];
              var symbol = { name: _this.getKeyLabel(property2), kind: _this.getSymbolKind(valueNode.type), range: range2, selectionRange: selectionRange2, children, detail: _this.getDetail(valueNode) };
              result2.push(symbol);
              toVisit.push({ result: children, node: valueNode });
            } else {
              limitExceeded = true;
            }
          }
        });
      }
    };
    while (nextToVisit < toVisit.length) {
      var next = toVisit[nextToVisit++];
      collectOutlineEntries(next.node, next.result);
    }
    if (limitExceeded && context && context.onResultLimitExceeded) {
      context.onResultLimitExceeded(resourceString);
    }
    return result;
  };
  JSONDocumentSymbols2.prototype.getSymbolKind = function(nodeType) {
    switch (nodeType) {
      case "object":
        return SymbolKind.Module;
      case "string":
        return SymbolKind.String;
      case "number":
        return SymbolKind.Number;
      case "array":
        return SymbolKind.Array;
      case "boolean":
        return SymbolKind.Boolean;
      default:
        return SymbolKind.Variable;
    }
  };
  JSONDocumentSymbols2.prototype.getKeyLabel = function(property) {
    var name = property.keyNode.value;
    if (name) {
      name = name.replace(/[\n]/g, "\u21B5");
    }
    if (name && name.trim()) {
      return name;
    }
    return '"' + name + '"';
  };
  JSONDocumentSymbols2.prototype.getDetail = function(node) {
    if (!node) {
      return void 0;
    }
    if (node.type === "boolean" || node.type === "number" || node.type === "null" || node.type === "string") {
      return String(node.value);
    } else {
      if (node.type === "array") {
        return node.children.length ? void 0 : "[]";
      } else if (node.type === "object") {
        return node.children.length ? void 0 : "{}";
      }
    }
    return void 0;
  };
  JSONDocumentSymbols2.prototype.findDocumentColors = function(document, doc, context) {
    return this.schemaService.getSchemaForResource(document.uri, doc).then(function(schema) {
      var result = [];
      if (schema) {
        var limit = context && typeof context.resultLimit === "number" ? context.resultLimit : Number.MAX_VALUE;
        var matchingSchemas = doc.getMatchingSchemas(schema.schema);
        var visitedNode = {};
        for (var _i = 0, matchingSchemas_1 = matchingSchemas; _i < matchingSchemas_1.length; _i++) {
          var s = matchingSchemas_1[_i];
          if (!s.inverted && s.schema && (s.schema.format === "color" || s.schema.format === "color-hex") && s.node && s.node.type === "string") {
            var nodeId = String(s.node.offset);
            if (!visitedNode[nodeId]) {
              var color = colorFromHex(getNodeValue2(s.node));
              if (color) {
                var range = getRange(document, s.node);
                result.push({ color, range });
              }
              visitedNode[nodeId] = true;
              limit--;
              if (limit <= 0) {
                if (context && context.onResultLimitExceeded) {
                  context.onResultLimitExceeded(document.uri);
                }
                return result;
              }
            }
          }
        }
      }
      return result;
    });
  };
  JSONDocumentSymbols2.prototype.getColorPresentations = function(document, doc, color, range) {
    var result = [];
    var red256 = Math.round(color.red * 255), green256 = Math.round(color.green * 255), blue256 = Math.round(color.blue * 255);
    function toTwoDigitHex(n) {
      var r = n.toString(16);
      return r.length !== 2 ? "0" + r : r;
    }
    var label;
    if (color.alpha === 1) {
      label = "#" + toTwoDigitHex(red256) + toTwoDigitHex(green256) + toTwoDigitHex(blue256);
    } else {
      label = "#" + toTwoDigitHex(red256) + toTwoDigitHex(green256) + toTwoDigitHex(blue256) + toTwoDigitHex(Math.round(color.alpha * 255));
    }
    result.push({ label, textEdit: TextEdit.replace(range, JSON.stringify(label)) });
    return result;
  };
  return JSONDocumentSymbols2;
}();
function getRange(document, node) {
  return Range.create(document.positionAt(node.offset), document.positionAt(node.offset + node.length));
}

// node_modules/vscode-json-languageservice/lib/esm/services/configuration.js
var localize6 = loadMessageBundle();
var schemaContributions = {
  schemaAssociations: [],
  schemas: {
    "http://json-schema.org/schema#": {
      $ref: "http://json-schema.org/draft-07/schema#"
    },
    "http://json-schema.org/draft-04/schema#": {
      "$schema": "http://json-schema.org/draft-04/schema#",
      "definitions": {
        "schemaArray": {
          "type": "array",
          "minItems": 1,
          "items": {
            "$ref": "#"
          }
        },
        "positiveInteger": {
          "type": "integer",
          "minimum": 0
        },
        "positiveIntegerDefault0": {
          "allOf": [
            {
              "$ref": "#/definitions/positiveInteger"
            },
            {
              "default": 0
            }
          ]
        },
        "simpleTypes": {
          "type": "string",
          "enum": [
            "array",
            "boolean",
            "integer",
            "null",
            "number",
            "object",
            "string"
          ]
        },
        "stringArray": {
          "type": "array",
          "items": {
            "type": "string"
          },
          "minItems": 1,
          "uniqueItems": true
        }
      },
      "type": "object",
      "properties": {
        "id": {
          "type": "string",
          "format": "uri"
        },
        "$schema": {
          "type": "string",
          "format": "uri"
        },
        "title": {
          "type": "string"
        },
        "description": {
          "type": "string"
        },
        "default": {},
        "multipleOf": {
          "type": "number",
          "minimum": 0,
          "exclusiveMinimum": true
        },
        "maximum": {
          "type": "number"
        },
        "exclusiveMaximum": {
          "type": "boolean",
          "default": false
        },
        "minimum": {
          "type": "number"
        },
        "exclusiveMinimum": {
          "type": "boolean",
          "default": false
        },
        "maxLength": {
          "allOf": [
            {
              "$ref": "#/definitions/positiveInteger"
            }
          ]
        },
        "minLength": {
          "allOf": [
            {
              "$ref": "#/definitions/positiveIntegerDefault0"
            }
          ]
        },
        "pattern": {
          "type": "string",
          "format": "regex"
        },
        "additionalItems": {
          "anyOf": [
            {
              "type": "boolean"
            },
            {
              "$ref": "#"
            }
          ],
          "default": {}
        },
        "items": {
          "anyOf": [
            {
              "$ref": "#"
            },
            {
              "$ref": "#/definitions/schemaArray"
            }
          ],
          "default": {}
        },
        "maxItems": {
          "allOf": [
            {
              "$ref": "#/definitions/positiveInteger"
            }
          ]
        },
        "minItems": {
          "allOf": [
            {
              "$ref": "#/definitions/positiveIntegerDefault0"
            }
          ]
        },
        "uniqueItems": {
          "type": "boolean",
          "default": false
        },
        "maxProperties": {
          "allOf": [
            {
              "$ref": "#/definitions/positiveInteger"
            }
          ]
        },
        "minProperties": {
          "allOf": [
            {
              "$ref": "#/definitions/positiveIntegerDefault0"
            }
          ]
        },
        "required": {
          "allOf": [
            {
              "$ref": "#/definitions/stringArray"
            }
          ]
        },
        "additionalProperties": {
          "anyOf": [
            {
              "type": "boolean"
            },
            {
              "$ref": "#"
            }
          ],
          "default": {}
        },
        "definitions": {
          "type": "object",
          "additionalProperties": {
            "$ref": "#"
          },
          "default": {}
        },
        "properties": {
          "type": "object",
          "additionalProperties": {
            "$ref": "#"
          },
          "default": {}
        },
        "patternProperties": {
          "type": "object",
          "additionalProperties": {
            "$ref": "#"
          },
          "default": {}
        },
        "dependencies": {
          "type": "object",
          "additionalProperties": {
            "anyOf": [
              {
                "$ref": "#"
              },
              {
                "$ref": "#/definitions/stringArray"
              }
            ]
          }
        },
        "enum": {
          "type": "array",
          "minItems": 1,
          "uniqueItems": true
        },
        "type": {
          "anyOf": [
            {
              "$ref": "#/definitions/simpleTypes"
            },
            {
              "type": "array",
              "items": {
                "$ref": "#/definitions/simpleTypes"
              },
              "minItems": 1,
              "uniqueItems": true
            }
          ]
        },
        "format": {
          "anyOf": [
            {
              "type": "string",
              "enum": [
                "date-time",
                "uri",
                "email",
                "hostname",
                "ipv4",
                "ipv6",
                "regex"
              ]
            },
            {
              "type": "string"
            }
          ]
        },
        "allOf": {
          "allOf": [
            {
              "$ref": "#/definitions/schemaArray"
            }
          ]
        },
        "anyOf": {
          "allOf": [
            {
              "$ref": "#/definitions/schemaArray"
            }
          ]
        },
        "oneOf": {
          "allOf": [
            {
              "$ref": "#/definitions/schemaArray"
            }
          ]
        },
        "not": {
          "allOf": [
            {
              "$ref": "#"
            }
          ]
        }
      },
      "dependencies": {
        "exclusiveMaximum": [
          "maximum"
        ],
        "exclusiveMinimum": [
          "minimum"
        ]
      },
      "default": {}
    },
    "http://json-schema.org/draft-07/schema#": {
      "definitions": {
        "schemaArray": {
          "type": "array",
          "minItems": 1,
          "items": { "$ref": "#" }
        },
        "nonNegativeInteger": {
          "type": "integer",
          "minimum": 0
        },
        "nonNegativeIntegerDefault0": {
          "allOf": [
            { "$ref": "#/definitions/nonNegativeInteger" },
            { "default": 0 }
          ]
        },
        "simpleTypes": {
          "enum": [
            "array",
            "boolean",
            "integer",
            "null",
            "number",
            "object",
            "string"
          ]
        },
        "stringArray": {
          "type": "array",
          "items": { "type": "string" },
          "uniqueItems": true,
          "default": []
        }
      },
      "type": ["object", "boolean"],
      "properties": {
        "$id": {
          "type": "string",
          "format": "uri-reference"
        },
        "$schema": {
          "type": "string",
          "format": "uri"
        },
        "$ref": {
          "type": "string",
          "format": "uri-reference"
        },
        "$comment": {
          "type": "string"
        },
        "title": {
          "type": "string"
        },
        "description": {
          "type": "string"
        },
        "default": true,
        "readOnly": {
          "type": "boolean",
          "default": false
        },
        "examples": {
          "type": "array",
          "items": true
        },
        "multipleOf": {
          "type": "number",
          "exclusiveMinimum": 0
        },
        "maximum": {
          "type": "number"
        },
        "exclusiveMaximum": {
          "type": "number"
        },
        "minimum": {
          "type": "number"
        },
        "exclusiveMinimum": {
          "type": "number"
        },
        "maxLength": { "$ref": "#/definitions/nonNegativeInteger" },
        "minLength": { "$ref": "#/definitions/nonNegativeIntegerDefault0" },
        "pattern": {
          "type": "string",
          "format": "regex"
        },
        "additionalItems": { "$ref": "#" },
        "items": {
          "anyOf": [
            { "$ref": "#" },
            { "$ref": "#/definitions/schemaArray" }
          ],
          "default": true
        },
        "maxItems": { "$ref": "#/definitions/nonNegativeInteger" },
        "minItems": { "$ref": "#/definitions/nonNegativeIntegerDefault0" },
        "uniqueItems": {
          "type": "boolean",
          "default": false
        },
        "contains": { "$ref": "#" },
        "maxProperties": { "$ref": "#/definitions/nonNegativeInteger" },
        "minProperties": { "$ref": "#/definitions/nonNegativeIntegerDefault0" },
        "required": { "$ref": "#/definitions/stringArray" },
        "additionalProperties": { "$ref": "#" },
        "definitions": {
          "type": "object",
          "additionalProperties": { "$ref": "#" },
          "default": {}
        },
        "properties": {
          "type": "object",
          "additionalProperties": { "$ref": "#" },
          "default": {}
        },
        "patternProperties": {
          "type": "object",
          "additionalProperties": { "$ref": "#" },
          "propertyNames": { "format": "regex" },
          "default": {}
        },
        "dependencies": {
          "type": "object",
          "additionalProperties": {
            "anyOf": [
              { "$ref": "#" },
              { "$ref": "#/definitions/stringArray" }
            ]
          }
        },
        "propertyNames": { "$ref": "#" },
        "const": true,
        "enum": {
          "type": "array",
          "items": true,
          "minItems": 1,
          "uniqueItems": true
        },
        "type": {
          "anyOf": [
            { "$ref": "#/definitions/simpleTypes" },
            {
              "type": "array",
              "items": { "$ref": "#/definitions/simpleTypes" },
              "minItems": 1,
              "uniqueItems": true
            }
          ]
        },
        "format": { "type": "string" },
        "contentMediaType": { "type": "string" },
        "contentEncoding": { "type": "string" },
        "if": { "$ref": "#" },
        "then": { "$ref": "#" },
        "else": { "$ref": "#" },
        "allOf": { "$ref": "#/definitions/schemaArray" },
        "anyOf": { "$ref": "#/definitions/schemaArray" },
        "oneOf": { "$ref": "#/definitions/schemaArray" },
        "not": { "$ref": "#" }
      },
      "default": true
    }
  }
};
var descriptions = {
  id: localize6("schema.json.id", "A unique identifier for the schema."),
  $schema: localize6("schema.json.$schema", "The schema to verify this document against."),
  title: localize6("schema.json.title", "A descriptive title of the element."),
  description: localize6("schema.json.description", "A long description of the element. Used in hover menus and suggestions."),
  default: localize6("schema.json.default", "A default value. Used by suggestions."),
  multipleOf: localize6("schema.json.multipleOf", "A number that should cleanly divide the current value (i.e. have no remainder)."),
  maximum: localize6("schema.json.maximum", "The maximum numerical value, inclusive by default."),
  exclusiveMaximum: localize6("schema.json.exclusiveMaximum", "Makes the maximum property exclusive."),
  minimum: localize6("schema.json.minimum", "The minimum numerical value, inclusive by default."),
  exclusiveMinimum: localize6("schema.json.exclusiveMininum", "Makes the minimum property exclusive."),
  maxLength: localize6("schema.json.maxLength", "The maximum length of a string."),
  minLength: localize6("schema.json.minLength", "The minimum length of a string."),
  pattern: localize6("schema.json.pattern", "A regular expression to match the string against. It is not implicitly anchored."),
  additionalItems: localize6("schema.json.additionalItems", "For arrays, only when items is set as an array. If it is a schema, then this schema validates items after the ones specified by the items array. If it is false, then additional items will cause validation to fail."),
  items: localize6("schema.json.items", "For arrays. Can either be a schema to validate every element against or an array of schemas to validate each item against in order (the first schema will validate the first element, the second schema will validate the second element, and so on."),
  maxItems: localize6("schema.json.maxItems", "The maximum number of items that can be inside an array. Inclusive."),
  minItems: localize6("schema.json.minItems", "The minimum number of items that can be inside an array. Inclusive."),
  uniqueItems: localize6("schema.json.uniqueItems", "If all of the items in the array must be unique. Defaults to false."),
  maxProperties: localize6("schema.json.maxProperties", "The maximum number of properties an object can have. Inclusive."),
  minProperties: localize6("schema.json.minProperties", "The minimum number of properties an object can have. Inclusive."),
  required: localize6("schema.json.required", "An array of strings that lists the names of all properties required on this object."),
  additionalProperties: localize6("schema.json.additionalProperties", "Either a schema or a boolean. If a schema, then used to validate all properties not matched by 'properties' or 'patternProperties'. If false, then any properties not matched by either will cause this schema to fail."),
  definitions: localize6("schema.json.definitions", "Not used for validation. Place subschemas here that you wish to reference inline with $ref."),
  properties: localize6("schema.json.properties", "A map of property names to schemas for each property."),
  patternProperties: localize6("schema.json.patternProperties", "A map of regular expressions on property names to schemas for matching properties."),
  dependencies: localize6("schema.json.dependencies", "A map of property names to either an array of property names or a schema. An array of property names means the property named in the key depends on the properties in the array being present in the object in order to be valid. If the value is a schema, then the schema is only applied to the object if the property in the key exists on the object."),
  enum: localize6("schema.json.enum", "The set of literal values that are valid."),
  type: localize6("schema.json.type", "Either a string of one of the basic schema types (number, integer, null, array, object, boolean, string) or an array of strings specifying a subset of those types."),
  format: localize6("schema.json.format", "Describes the format expected for the value."),
  allOf: localize6("schema.json.allOf", "An array of schemas, all of which must match."),
  anyOf: localize6("schema.json.anyOf", "An array of schemas, where at least one must match."),
  oneOf: localize6("schema.json.oneOf", "An array of schemas, exactly one of which must match."),
  not: localize6("schema.json.not", "A schema which must not match."),
  $id: localize6("schema.json.$id", "A unique identifier for the schema."),
  $ref: localize6("schema.json.$ref", "Reference a definition hosted on any location."),
  $comment: localize6("schema.json.$comment", "Comments from schema authors to readers or maintainers of the schema."),
  readOnly: localize6("schema.json.readOnly", "Indicates that the value of the instance is managed exclusively by the owning authority."),
  examples: localize6("schema.json.examples", "Sample JSON values associated with a particular schema, for the purpose of illustrating usage."),
  contains: localize6("schema.json.contains", 'An array instance is valid against "contains" if at least one of its elements is valid against the given schema.'),
  propertyNames: localize6("schema.json.propertyNames", "If the instance is an object, this keyword validates if every property name in the instance validates against the provided schema."),
  const: localize6("schema.json.const", "An instance validates successfully against this keyword if its value is equal to the value of the keyword."),
  contentMediaType: localize6("schema.json.contentMediaType", "Describes the media type of a string property."),
  contentEncoding: localize6("schema.json.contentEncoding", "Describes the content encoding of a string property."),
  if: localize6("schema.json.if", 'The validation outcome of the "if" subschema controls which of the "then" or "else" keywords are evaluated.'),
  then: localize6("schema.json.then", 'The "if" subschema is used for validation when the "if" subschema succeeds.'),
  else: localize6("schema.json.else", 'The "else" subschema is used for validation when the "if" subschema fails.')
};
for (schemaName in schemaContributions.schemas) {
  schema = schemaContributions.schemas[schemaName];
  for (property in schema.properties) {
    propertyObject = schema.properties[property];
    if (typeof propertyObject === "boolean") {
      propertyObject = schema.properties[property] = {};
    }
    description = descriptions[property];
    if (description) {
      propertyObject["description"] = description;
    } else {
      console.log(property + ": localize('schema.json." + property + `', "")`);
    }
  }
}
var schema;
var propertyObject;
var description;
var property;
var schemaName;

// node_modules/vscode-json-languageservice/lib/esm/services/jsonFolding.js
import { createScanner as createScanner3 } from "jsonc-parser";

// node_modules/vscode-json-languageservice/lib/esm/services/jsonSelectionRanges.js
import { createScanner as createScanner4 } from "jsonc-parser";

// node_modules/vscode-json-languageservice/lib/esm/jsonLanguageService.js
import { format as formatJSON } from "jsonc-parser";

// node_modules/vscode-json-languageservice/lib/esm/services/jsonLinks.js
function findLinks(document, doc) {
  var links = [];
  doc.visit(function(node) {
    var _a;
    if (node.type === "property" && node.keyNode.value === "$ref" && ((_a = node.valueNode) === null || _a === void 0 ? void 0 : _a.type) === "string") {
      var path5 = node.valueNode.value;
      var targetNode = findTargetNode(doc, path5);
      if (targetNode) {
        var targetPos = document.positionAt(targetNode.offset);
        links.push({
          target: document.uri + "#" + (targetPos.line + 1) + "," + (targetPos.character + 1),
          range: createRange(document, node.valueNode)
        });
      }
    }
    return true;
  });
  return Promise.resolve(links);
}
function createRange(document, node) {
  return Range.create(document.positionAt(node.offset + 1), document.positionAt(node.offset + node.length - 1));
}
function findTargetNode(doc, path5) {
  var tokens = parseJSONPointer(path5);
  if (!tokens) {
    return null;
  }
  return findNode(tokens, doc.root);
}
function findNode(pointer, node) {
  if (!node) {
    return null;
  }
  if (pointer.length === 0) {
    return node;
  }
  var token = pointer.shift();
  if (node && node.type === "object") {
    var propertyNode = node.properties.find(function(propertyNode2) {
      return propertyNode2.keyNode.value === token;
    });
    if (!propertyNode) {
      return null;
    }
    return findNode(pointer, propertyNode.valueNode);
  } else if (node && node.type === "array") {
    if (token.match(/^(0|[1-9][0-9]*)$/)) {
      var index = Number.parseInt(token);
      var arrayItem = node.items[index];
      if (!arrayItem) {
        return null;
      }
      return findNode(pointer, arrayItem);
    }
  }
  return null;
}
function parseJSONPointer(path5) {
  if (path5 === "#") {
    return [];
  }
  if (path5[0] !== "#" || path5[1] !== "/") {
    return null;
  }
  return path5.substring(2).split(/\//).map(unescape);
}
function unescape(str) {
  return str.replace(/~1/g, "/").replace(/~0/g, "~");
}

// node_modules/yaml-language-server/lib/esm/languageservice/parser/jsonParser07.js
import { DiagnosticSeverity as DiagnosticSeverity2, Range as Range2 } from "vscode-languageserver-types";
import { Diagnostic as Diagnostic2 } from "vscode-languageserver-types";

// node_modules/yaml-language-server/lib/esm/languageservice/utils/arrUtils.js
function matchOffsetToDocument(offset, jsonDocuments) {
  for (const jsonDoc of jsonDocuments.documents) {
    if (jsonDoc.internalDocument && jsonDoc.internalDocument.range[0] <= offset && jsonDoc.internalDocument.range[2] >= offset) {
      return jsonDoc;
    }
  }
  if (jsonDocuments.documents.length === 1) {
    return jsonDocuments.documents[0];
  }
  return null;
}
function filterInvalidCustomTags(customTags) {
  const validCustomTags = ["mapping", "scalar", "sequence"];
  return customTags.filter((tag) => {
    if (typeof tag === "string") {
      const typeInfo = tag.split(" ");
      const type = typeInfo[1] && typeInfo[1].toLowerCase() || "scalar";
      if (type === "map") {
        return false;
      }
      return validCustomTags.indexOf(type) !== -1;
    }
    return false;
  });
}
function isArrayEqual(fst, snd) {
  if (!snd || !fst) {
    return false;
  }
  if (snd.length !== fst.length) {
    return false;
  }
  for (let index = fst.length - 1; index >= 0; index--) {
    if (fst[index] !== snd[index]) {
      return false;
    }
  }
  return true;
}

// node_modules/yaml-language-server/lib/esm/languageservice/parser/jsonParser07.js
var localize7 = loadMessageBundle();
var formats2 = {
  "color-hex": {
    errorMessage: localize7("colorHexFormatWarning", "Invalid color format. Use #RGB, #RGBA, #RRGGBB or #RRGGBBAA."),
    pattern: /^#([0-9A-Fa-f]{3,4}|([0-9A-Fa-f]{2}){3,4})$/
  },
  "date-time": {
    errorMessage: localize7("dateTimeFormatWarning", "String is not a RFC3339 date-time."),
    pattern: /^(\d{4})-(0[1-9]|1[0-2])-(0[1-9]|[12][0-9]|3[01])T([01][0-9]|2[0-3]):([0-5][0-9]):([0-5][0-9]|60)(\.[0-9]+)?(Z|(\+|-)([01][0-9]|2[0-3]):([0-5][0-9]))$/i
  },
  date: {
    errorMessage: localize7("dateFormatWarning", "String is not a RFC3339 date."),
    pattern: /^(\d{4})-(0[1-9]|1[0-2])-(0[1-9]|[12][0-9]|3[01])$/i
  },
  time: {
    errorMessage: localize7("timeFormatWarning", "String is not a RFC3339 time."),
    pattern: /^([01][0-9]|2[0-3]):([0-5][0-9]):([0-5][0-9]|60)(\.[0-9]+)?(Z|(\+|-)([01][0-9]|2[0-3]):([0-5][0-9]))$/i
  },
  email: {
    errorMessage: localize7("emailFormatWarning", "String is not an e-mail address."),
    pattern: /^(([^<>()[\]\\.,;:\s@"]+(\.[^<>()[\]\\.,;:\s@"]+)*)|(".+"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$/
  }
};
var YAML_SOURCE = "YAML";
var YAML_SCHEMA_PREFIX = "yaml-schema: ";
var ProblemType;
(function(ProblemType2) {
  ProblemType2["missingRequiredPropWarning"] = "missingRequiredPropWarning";
  ProblemType2["typeMismatchWarning"] = "typeMismatchWarning";
  ProblemType2["constWarning"] = "constWarning";
})(ProblemType || (ProblemType = {}));
var ProblemTypeMessages = {
  [ProblemType.missingRequiredPropWarning]: 'Missing property "{0}".',
  [ProblemType.typeMismatchWarning]: 'Incorrect type. Expected "{0}".',
  [ProblemType.constWarning]: "Value must be {0}."
};
var ASTNodeImpl2 = class {
  constructor(parent, internalNode, offset, length) {
    this.offset = offset;
    this.length = length;
    this.parent = parent;
    this.internalNode = internalNode;
  }
  getNodeFromOffsetEndInclusive(offset) {
    const collector = [];
    const findNode2 = (node) => {
      if (offset >= node.offset && offset <= node.offset + node.length) {
        const children = node.children;
        for (let i = 0; i < children.length && children[i].offset <= offset; i++) {
          const item = findNode2(children[i]);
          if (item) {
            collector.push(item);
          }
        }
        return node;
      }
      return null;
    };
    const foundNode = findNode2(this);
    let currMinDist = Number.MAX_VALUE;
    let currMinNode = null;
    for (const currNode of collector) {
      const minDist = currNode.length + currNode.offset - offset + (offset - currNode.offset);
      if (minDist < currMinDist) {
        currMinNode = currNode;
        currMinDist = minDist;
      }
    }
    return currMinNode || foundNode;
  }
  get children() {
    return [];
  }
  toString() {
    return "type: " + this.type + " (" + this.offset + "/" + this.length + ")" + (this.parent ? " parent: {" + this.parent.toString() + "}" : "");
  }
};
var NullASTNodeImpl2 = class extends ASTNodeImpl2 {
  constructor(parent, internalNode, offset, length) {
    super(parent, internalNode, offset, length);
    this.type = "null";
    this.value = null;
  }
};
var BooleanASTNodeImpl2 = class extends ASTNodeImpl2 {
  constructor(parent, internalNode, boolValue, offset, length) {
    super(parent, internalNode, offset, length);
    this.type = "boolean";
    this.value = boolValue;
  }
};
var ArrayASTNodeImpl2 = class extends ASTNodeImpl2 {
  constructor(parent, internalNode, offset, length) {
    super(parent, internalNode, offset, length);
    this.type = "array";
    this.items = [];
  }
  get children() {
    return this.items;
  }
};
var NumberASTNodeImpl2 = class extends ASTNodeImpl2 {
  constructor(parent, internalNode, offset, length) {
    super(parent, internalNode, offset, length);
    this.type = "number";
    this.isInteger = true;
    this.value = Number.NaN;
  }
};
var StringASTNodeImpl2 = class extends ASTNodeImpl2 {
  constructor(parent, internalNode, offset, length) {
    super(parent, internalNode, offset, length);
    this.type = "string";
    this.value = "";
  }
};
var PropertyASTNodeImpl2 = class extends ASTNodeImpl2 {
  constructor(parent, internalNode, offset, length) {
    super(parent, internalNode, offset, length);
    this.type = "property";
    this.colonOffset = -1;
  }
  get children() {
    return this.valueNode ? [this.keyNode, this.valueNode] : [this.keyNode];
  }
};
var ObjectASTNodeImpl2 = class extends ASTNodeImpl2 {
  constructor(parent, internalNode, offset, length) {
    super(parent, internalNode, offset, length);
    this.type = "object";
    this.properties = [];
  }
  get children() {
    return this.properties;
  }
};
function asSchema2(schema) {
  if (isBoolean2(schema)) {
    return schema ? {} : { not: {} };
  }
  return schema;
}
var EnumMatch2;
(function(EnumMatch3) {
  EnumMatch3[EnumMatch3["Key"] = 0] = "Key";
  EnumMatch3[EnumMatch3["Enum"] = 1] = "Enum";
})(EnumMatch2 || (EnumMatch2 = {}));
var SchemaCollector2 = class {
  constructor(focusOffset = -1, exclude = null) {
    this.focusOffset = focusOffset;
    this.exclude = exclude;
    this.schemas = [];
  }
  add(schema) {
    this.schemas.push(schema);
  }
  merge(other) {
    this.schemas.push(...other.schemas);
  }
  include(node) {
    return (this.focusOffset === -1 || contains2(node, this.focusOffset)) && node !== this.exclude;
  }
  newSub() {
    return new SchemaCollector2(-1, this.exclude);
  }
};
var NoOpSchemaCollector2 = class {
  constructor() {
  }
  get schemas() {
    return [];
  }
  add(schema) {
  }
  merge(other) {
  }
  include(node) {
    return true;
  }
  newSub() {
    return this;
  }
};
NoOpSchemaCollector2.instance = new NoOpSchemaCollector2();
var ValidationResult2 = class {
  constructor(isKubernetes) {
    this.problems = [];
    this.propertiesMatches = 0;
    this.propertiesValueMatches = 0;
    this.primaryValueMatches = 0;
    this.enumValueMatch = false;
    if (isKubernetes) {
      this.enumValues = [];
    } else {
      this.enumValues = null;
    }
  }
  hasProblems() {
    return !!this.problems.length;
  }
  mergeAll(validationResults) {
    for (const validationResult of validationResults) {
      this.merge(validationResult);
    }
  }
  merge(validationResult) {
    this.problems = this.problems.concat(validationResult.problems);
  }
  mergeEnumValues(validationResult) {
    if (!this.enumValueMatch && !validationResult.enumValueMatch && this.enumValues && validationResult.enumValues) {
      this.enumValues = this.enumValues.concat(validationResult.enumValues);
      for (const error of this.problems) {
        if (error.code === ErrorCode.EnumValueMismatch) {
          error.message = localize7("enumWarning", "Value is not accepted. Valid values: {0}.", [...new Set(this.enumValues)].map((v) => {
            return JSON.stringify(v);
          }).join(", "));
        }
      }
    }
  }
  mergeWarningGeneric(subValidationResult, problemTypesToMerge) {
    var _a, _b;
    if ((_a = this.problems) === null || _a === void 0 ? void 0 : _a.length) {
      for (const problemType of problemTypesToMerge) {
        const bestResults = this.problems.filter((p) => p.problemType === problemType);
        for (const bestResult of bestResults) {
          const mergingResult = (_b = subValidationResult.problems) === null || _b === void 0 ? void 0 : _b.find((p) => p.problemType === problemType && bestResult.location.offset === p.location.offset && (problemType !== ProblemType.missingRequiredPropWarning || isArrayEqual(p.problemArgs, bestResult.problemArgs)));
          if (mergingResult) {
            if (mergingResult.problemArgs.length) {
              mergingResult.problemArgs.filter((p) => !bestResult.problemArgs.includes(p)).forEach((p) => bestResult.problemArgs.push(p));
              bestResult.message = getWarningMessage(bestResult.problemType, bestResult.problemArgs);
            }
            this.mergeSources(mergingResult, bestResult);
          }
        }
      }
    }
  }
  mergePropertyMatch(propertyValidationResult) {
    this.merge(propertyValidationResult);
    this.propertiesMatches++;
    if (propertyValidationResult.enumValueMatch || !propertyValidationResult.hasProblems() && propertyValidationResult.propertiesMatches) {
      this.propertiesValueMatches++;
    }
    if (propertyValidationResult.enumValueMatch && propertyValidationResult.enumValues) {
      this.primaryValueMatches++;
    }
  }
  mergeSources(mergingResult, bestResult) {
    const mergingSource = mergingResult.source.replace(YAML_SCHEMA_PREFIX, "");
    if (!bestResult.source.includes(mergingSource)) {
      bestResult.source = bestResult.source + " | " + mergingSource;
    }
    if (!bestResult.schemaUri.includes(mergingResult.schemaUri[0])) {
      bestResult.schemaUri = bestResult.schemaUri.concat(mergingResult.schemaUri);
    }
  }
  compareGeneric(other) {
    const hasProblems = this.hasProblems();
    if (hasProblems !== other.hasProblems()) {
      return hasProblems ? -1 : 1;
    }
    if (this.enumValueMatch !== other.enumValueMatch) {
      return other.enumValueMatch ? -1 : 1;
    }
    if (this.propertiesValueMatches !== other.propertiesValueMatches) {
      return this.propertiesValueMatches - other.propertiesValueMatches;
    }
    if (this.primaryValueMatches !== other.primaryValueMatches) {
      return this.primaryValueMatches - other.primaryValueMatches;
    }
    return this.propertiesMatches - other.propertiesMatches;
  }
  compareKubernetes(other) {
    const hasProblems = this.hasProblems();
    if (this.propertiesMatches !== other.propertiesMatches) {
      return this.propertiesMatches - other.propertiesMatches;
    }
    if (this.enumValueMatch !== other.enumValueMatch) {
      return other.enumValueMatch ? -1 : 1;
    }
    if (this.primaryValueMatches !== other.primaryValueMatches) {
      return this.primaryValueMatches - other.primaryValueMatches;
    }
    if (this.propertiesValueMatches !== other.propertiesValueMatches) {
      return this.propertiesValueMatches - other.propertiesValueMatches;
    }
    if (hasProblems !== other.hasProblems()) {
      return hasProblems ? -1 : 1;
    }
    return this.propertiesMatches - other.propertiesMatches;
  }
};
function getNodeValue4(node) {
  return getNodeValue3(node);
}
function contains2(node, offset, includeRightBound = false) {
  return offset >= node.offset && offset <= node.offset + node.length || includeRightBound && offset === node.offset + node.length;
}
var JSONDocument2 = class {
  constructor(root, syntaxErrors = [], comments = []) {
    this.root = root;
    this.syntaxErrors = syntaxErrors;
    this.comments = comments;
  }
  getNodeFromOffset(offset, includeRightBound = false) {
    if (this.root) {
      return findNodeAtOffset2(this.root, offset, includeRightBound);
    }
    return void 0;
  }
  getNodeFromOffsetEndInclusive(offset) {
    return this.root && this.root.getNodeFromOffsetEndInclusive(offset);
  }
  visit(visitor) {
    if (this.root) {
      const doVisit = (node) => {
        let ctn = visitor(node);
        const children = node.children;
        if (Array.isArray(children)) {
          for (let i = 0; i < children.length && ctn; i++) {
            ctn = doVisit(children[i]);
          }
        }
        return ctn;
      };
      doVisit(this.root);
    }
  }
  validate(textDocument, schema) {
    if (this.root && schema) {
      const validationResult = new ValidationResult2(this.isKubernetes);
      validate2(this.root, schema, schema, validationResult, NoOpSchemaCollector2.instance, {
        isKubernetes: this.isKubernetes,
        disableAdditionalProperties: this.disableAdditionalProperties
      });
      return validationResult.problems.map((p) => {
        const range = Range2.create(textDocument.positionAt(p.location.offset), textDocument.positionAt(p.location.offset + p.location.length));
        const diagnostic = Diagnostic2.create(range, p.message, p.severity, p.code ? p.code : ErrorCode.Undefined, p.source);
        diagnostic.data = { schemaUri: p.schemaUri };
        return diagnostic;
      });
    }
    return null;
  }
  getMatchingSchemas(schema, focusOffset = -1, exclude = null) {
    const matchingSchemas = new SchemaCollector2(focusOffset, exclude);
    if (this.root && schema) {
      validate2(this.root, schema, schema, new ValidationResult2(this.isKubernetes), matchingSchemas, {
        isKubernetes: this.isKubernetes,
        disableAdditionalProperties: this.disableAdditionalProperties
      });
    }
    return matchingSchemas.schemas;
  }
};
function validate2(node, schema, originalSchema, validationResult, matchingSchemas, options) {
  const { isKubernetes } = options;
  if (!node || !matchingSchemas.include(node)) {
    return;
  }
  if (!schema.url) {
    schema.url = originalSchema.url;
  }
  if (!schema.title) {
    schema.title = originalSchema.title;
  }
  switch (node.type) {
    case "object":
      _validateObjectNode(node, schema, validationResult, matchingSchemas);
      break;
    case "array":
      _validateArrayNode(node, schema, validationResult, matchingSchemas);
      break;
    case "string":
      _validateStringNode(node, schema, validationResult);
      break;
    case "number":
      _validateNumberNode(node, schema, validationResult);
      break;
    case "property":
      return validate2(node.valueNode, schema, schema, validationResult, matchingSchemas, options);
  }
  _validateNode();
  matchingSchemas.add({ node, schema });
  function _validateNode() {
    function matchesType(type) {
      return node.type === type || type === "integer" && node.type === "number" && node.isInteger;
    }
    if (Array.isArray(schema.type)) {
      if (!schema.type.some(matchesType)) {
        validationResult.problems.push({
          location: { offset: node.offset, length: node.length },
          severity: DiagnosticSeverity2.Warning,
          message: schema.errorMessage || localize7("typeArrayMismatchWarning", "Incorrect type. Expected one of {0}.", schema.type.join(", ")),
          source: getSchemaSource(schema, originalSchema),
          schemaUri: getSchemaUri(schema, originalSchema)
        });
      }
    } else if (schema.type) {
      if (!matchesType(schema.type)) {
        const schemaType = schema.type === "object" ? getSchemaTypeName(schema) : schema.type;
        validationResult.problems.push({
          location: { offset: node.offset, length: node.length },
          severity: DiagnosticSeverity2.Warning,
          message: schema.errorMessage || getWarningMessage(ProblemType.typeMismatchWarning, [schemaType]),
          source: getSchemaSource(schema, originalSchema),
          schemaUri: getSchemaUri(schema, originalSchema),
          problemType: ProblemType.typeMismatchWarning,
          problemArgs: [schemaType]
        });
      }
    }
    if (Array.isArray(schema.allOf)) {
      for (const subSchemaRef of schema.allOf) {
        validate2(node, asSchema2(subSchemaRef), schema, validationResult, matchingSchemas, options);
      }
    }
    const notSchema = asSchema2(schema.not);
    if (notSchema) {
      const subValidationResult = new ValidationResult2(isKubernetes);
      const subMatchingSchemas = matchingSchemas.newSub();
      validate2(node, notSchema, schema, subValidationResult, subMatchingSchemas, options);
      if (!subValidationResult.hasProblems()) {
        validationResult.problems.push({
          location: { offset: node.offset, length: node.length },
          severity: DiagnosticSeverity2.Warning,
          message: localize7("notSchemaWarning", "Matches a schema that is not allowed."),
          source: getSchemaSource(schema, originalSchema),
          schemaUri: getSchemaUri(schema, originalSchema)
        });
      }
      for (const ms of subMatchingSchemas.schemas) {
        ms.inverted = !ms.inverted;
        matchingSchemas.add(ms);
      }
    }
    const testAlternatives = (alternatives, maxOneMatch) => {
      const matches = [];
      let bestMatch = null;
      for (const subSchemaRef of alternatives) {
        const subSchema = asSchema2(subSchemaRef);
        const subValidationResult = new ValidationResult2(isKubernetes);
        const subMatchingSchemas = matchingSchemas.newSub();
        validate2(node, subSchema, schema, subValidationResult, subMatchingSchemas, options);
        if (!subValidationResult.hasProblems()) {
          matches.push(subSchema);
        }
        if (!bestMatch) {
          bestMatch = {
            schema: subSchema,
            validationResult: subValidationResult,
            matchingSchemas: subMatchingSchemas
          };
        } else if (isKubernetes) {
          bestMatch = alternativeComparison(subValidationResult, bestMatch, subSchema, subMatchingSchemas);
        } else {
          bestMatch = genericComparison(maxOneMatch, subValidationResult, bestMatch, subSchema, subMatchingSchemas);
        }
      }
      if (matches.length > 1 && maxOneMatch) {
        validationResult.problems.push({
          location: { offset: node.offset, length: 1 },
          severity: DiagnosticSeverity2.Warning,
          message: localize7("oneOfWarning", "Matches multiple schemas when only one must validate."),
          source: getSchemaSource(schema, originalSchema),
          schemaUri: getSchemaUri(schema, originalSchema)
        });
      }
      if (bestMatch !== null) {
        validationResult.merge(bestMatch.validationResult);
        validationResult.propertiesMatches += bestMatch.validationResult.propertiesMatches;
        validationResult.propertiesValueMatches += bestMatch.validationResult.propertiesValueMatches;
        matchingSchemas.merge(bestMatch.matchingSchemas);
      }
      return matches.length;
    };
    if (Array.isArray(schema.anyOf)) {
      testAlternatives(schema.anyOf, false);
    }
    if (Array.isArray(schema.oneOf)) {
      testAlternatives(schema.oneOf, true);
    }
    const testBranch = (schema2, originalSchema2) => {
      const subValidationResult = new ValidationResult2(isKubernetes);
      const subMatchingSchemas = matchingSchemas.newSub();
      validate2(node, asSchema2(schema2), originalSchema2, subValidationResult, subMatchingSchemas, options);
      validationResult.merge(subValidationResult);
      validationResult.propertiesMatches += subValidationResult.propertiesMatches;
      validationResult.propertiesValueMatches += subValidationResult.propertiesValueMatches;
      matchingSchemas.merge(subMatchingSchemas);
    };
    const testCondition = (ifSchema2, originalSchema2, thenSchema, elseSchema) => {
      const subSchema = asSchema2(ifSchema2);
      const subValidationResult = new ValidationResult2(isKubernetes);
      const subMatchingSchemas = matchingSchemas.newSub();
      validate2(node, subSchema, originalSchema2, subValidationResult, subMatchingSchemas, options);
      matchingSchemas.merge(subMatchingSchemas);
      if (!subValidationResult.hasProblems()) {
        if (thenSchema) {
          testBranch(thenSchema, originalSchema2);
        }
      } else if (elseSchema) {
        testBranch(elseSchema, originalSchema2);
      }
    };
    const ifSchema = asSchema2(schema.if);
    if (ifSchema) {
      testCondition(ifSchema, schema, asSchema2(schema.then), asSchema2(schema.else));
    }
    if (Array.isArray(schema.enum)) {
      const val = getNodeValue4(node);
      let enumValueMatch = false;
      for (const e of schema.enum) {
        if (equals2(val, e)) {
          enumValueMatch = true;
          break;
        }
      }
      validationResult.enumValues = schema.enum;
      validationResult.enumValueMatch = enumValueMatch;
      if (!enumValueMatch) {
        validationResult.problems.push({
          location: { offset: node.offset, length: node.length },
          severity: DiagnosticSeverity2.Warning,
          code: ErrorCode.EnumValueMismatch,
          message: schema.errorMessage || localize7("enumWarning", "Value is not accepted. Valid values: {0}.", schema.enum.map((v) => {
            return JSON.stringify(v);
          }).join(", ")),
          source: getSchemaSource(schema, originalSchema),
          schemaUri: getSchemaUri(schema, originalSchema)
        });
      }
    }
    if (isDefined2(schema.const)) {
      const val = getNodeValue4(node);
      if (!equals2(val, schema.const)) {
        validationResult.problems.push({
          location: { offset: node.offset, length: node.length },
          severity: DiagnosticSeverity2.Warning,
          code: ErrorCode.EnumValueMismatch,
          problemType: ProblemType.constWarning,
          message: schema.errorMessage || getWarningMessage(ProblemType.constWarning, [JSON.stringify(schema.const)]),
          source: getSchemaSource(schema, originalSchema),
          schemaUri: getSchemaUri(schema, originalSchema),
          problemArgs: [JSON.stringify(schema.const)]
        });
        validationResult.enumValueMatch = false;
      } else {
        validationResult.enumValueMatch = true;
      }
      validationResult.enumValues = [schema.const];
    }
    if (schema.deprecationMessage && node.parent) {
      validationResult.problems.push({
        location: { offset: node.parent.offset, length: node.parent.length },
        severity: DiagnosticSeverity2.Warning,
        message: schema.deprecationMessage,
        source: getSchemaSource(schema, originalSchema),
        schemaUri: getSchemaUri(schema, originalSchema)
      });
    }
  }
  function _validateNumberNode(node2, schema2, validationResult2) {
    const val = node2.value;
    if (isNumber2(schema2.multipleOf)) {
      if (val % schema2.multipleOf !== 0) {
        validationResult2.problems.push({
          location: { offset: node2.offset, length: node2.length },
          severity: DiagnosticSeverity2.Warning,
          message: localize7("multipleOfWarning", "Value is not divisible by {0}.", schema2.multipleOf),
          source: getSchemaSource(schema2, originalSchema),
          schemaUri: getSchemaUri(schema2, originalSchema)
        });
      }
    }
    function getExclusiveLimit(limit, exclusive) {
      if (isNumber2(exclusive)) {
        return exclusive;
      }
      if (isBoolean2(exclusive) && exclusive) {
        return limit;
      }
      return void 0;
    }
    function getLimit(limit, exclusive) {
      if (!isBoolean2(exclusive) || !exclusive) {
        return limit;
      }
      return void 0;
    }
    const exclusiveMinimum = getExclusiveLimit(schema2.minimum, schema2.exclusiveMinimum);
    if (isNumber2(exclusiveMinimum) && val <= exclusiveMinimum) {
      validationResult2.problems.push({
        location: { offset: node2.offset, length: node2.length },
        severity: DiagnosticSeverity2.Warning,
        message: localize7("exclusiveMinimumWarning", "Value is below the exclusive minimum of {0}.", exclusiveMinimum),
        source: getSchemaSource(schema2, originalSchema),
        schemaUri: getSchemaUri(schema2, originalSchema)
      });
    }
    const exclusiveMaximum = getExclusiveLimit(schema2.maximum, schema2.exclusiveMaximum);
    if (isNumber2(exclusiveMaximum) && val >= exclusiveMaximum) {
      validationResult2.problems.push({
        location: { offset: node2.offset, length: node2.length },
        severity: DiagnosticSeverity2.Warning,
        message: localize7("exclusiveMaximumWarning", "Value is above the exclusive maximum of {0}.", exclusiveMaximum),
        source: getSchemaSource(schema2, originalSchema),
        schemaUri: getSchemaUri(schema2, originalSchema)
      });
    }
    const minimum = getLimit(schema2.minimum, schema2.exclusiveMinimum);
    if (isNumber2(minimum) && val < minimum) {
      validationResult2.problems.push({
        location: { offset: node2.offset, length: node2.length },
        severity: DiagnosticSeverity2.Warning,
        message: localize7("minimumWarning", "Value is below the minimum of {0}.", minimum),
        source: getSchemaSource(schema2, originalSchema),
        schemaUri: getSchemaUri(schema2, originalSchema)
      });
    }
    const maximum = getLimit(schema2.maximum, schema2.exclusiveMaximum);
    if (isNumber2(maximum) && val > maximum) {
      validationResult2.problems.push({
        location: { offset: node2.offset, length: node2.length },
        severity: DiagnosticSeverity2.Warning,
        message: localize7("maximumWarning", "Value is above the maximum of {0}.", maximum),
        source: getSchemaSource(schema2, originalSchema),
        schemaUri: getSchemaUri(schema2, originalSchema)
      });
    }
  }
  function _validateStringNode(node2, schema2, validationResult2) {
    if (isNumber2(schema2.minLength) && node2.value.length < schema2.minLength) {
      validationResult2.problems.push({
        location: { offset: node2.offset, length: node2.length },
        severity: DiagnosticSeverity2.Warning,
        message: localize7("minLengthWarning", "String is shorter than the minimum length of {0}.", schema2.minLength),
        source: getSchemaSource(schema2, originalSchema),
        schemaUri: getSchemaUri(schema2, originalSchema)
      });
    }
    if (isNumber2(schema2.maxLength) && node2.value.length > schema2.maxLength) {
      validationResult2.problems.push({
        location: { offset: node2.offset, length: node2.length },
        severity: DiagnosticSeverity2.Warning,
        message: localize7("maxLengthWarning", "String is longer than the maximum length of {0}.", schema2.maxLength),
        source: getSchemaSource(schema2, originalSchema),
        schemaUri: getSchemaUri(schema2, originalSchema)
      });
    }
    if (isString2(schema2.pattern)) {
      const regex = safeCreateUnicodeRegExp(schema2.pattern);
      if (!regex.test(node2.value)) {
        validationResult2.problems.push({
          location: { offset: node2.offset, length: node2.length },
          severity: DiagnosticSeverity2.Warning,
          message: schema2.patternErrorMessage || schema2.errorMessage || localize7("patternWarning", 'String does not match the pattern of "{0}".', schema2.pattern),
          source: getSchemaSource(schema2, originalSchema),
          schemaUri: getSchemaUri(schema2, originalSchema)
        });
      }
    }
    if (schema2.format) {
      switch (schema2.format) {
        case "uri":
        case "uri-reference":
          {
            let errorMessage;
            if (!node2.value) {
              errorMessage = localize7("uriEmpty", "URI expected.");
            } else {
              try {
                const uri = URI.parse(node2.value);
                if (!uri.scheme && schema2.format === "uri") {
                  errorMessage = localize7("uriSchemeMissing", "URI with a scheme is expected.");
                }
              } catch (e) {
                errorMessage = e.message;
              }
            }
            if (errorMessage) {
              validationResult2.problems.push({
                location: { offset: node2.offset, length: node2.length },
                severity: DiagnosticSeverity2.Warning,
                message: schema2.patternErrorMessage || schema2.errorMessage || localize7("uriFormatWarning", "String is not a URI: {0}", errorMessage),
                source: getSchemaSource(schema2, originalSchema),
                schemaUri: getSchemaUri(schema2, originalSchema)
              });
            }
          }
          break;
        case "color-hex":
        case "date-time":
        case "date":
        case "time":
        case "email":
          {
            const format3 = formats2[schema2.format];
            if (!node2.value || !format3.pattern.exec(node2.value)) {
              validationResult2.problems.push({
                location: { offset: node2.offset, length: node2.length },
                severity: DiagnosticSeverity2.Warning,
                message: schema2.patternErrorMessage || schema2.errorMessage || format3.errorMessage,
                source: getSchemaSource(schema2, originalSchema),
                schemaUri: getSchemaUri(schema2, originalSchema)
              });
            }
          }
          break;
        default:
      }
    }
  }
  function _validateArrayNode(node2, schema2, validationResult2, matchingSchemas2) {
    if (Array.isArray(schema2.items)) {
      const subSchemas = schema2.items;
      for (let index = 0; index < subSchemas.length; index++) {
        const subSchemaRef = subSchemas[index];
        const subSchema = asSchema2(subSchemaRef);
        const itemValidationResult = new ValidationResult2(isKubernetes);
        const item = node2.items[index];
        if (item) {
          validate2(item, subSchema, schema2, itemValidationResult, matchingSchemas2, options);
          validationResult2.mergePropertyMatch(itemValidationResult);
          validationResult2.mergeEnumValues(itemValidationResult);
        } else if (node2.items.length >= subSchemas.length) {
          validationResult2.propertiesValueMatches++;
        }
      }
      if (node2.items.length > subSchemas.length) {
        if (typeof schema2.additionalItems === "object") {
          for (let i = subSchemas.length; i < node2.items.length; i++) {
            const itemValidationResult = new ValidationResult2(isKubernetes);
            validate2(node2.items[i], schema2.additionalItems, schema2, itemValidationResult, matchingSchemas2, options);
            validationResult2.mergePropertyMatch(itemValidationResult);
            validationResult2.mergeEnumValues(itemValidationResult);
          }
        } else if (schema2.additionalItems === false) {
          validationResult2.problems.push({
            location: { offset: node2.offset, length: node2.length },
            severity: DiagnosticSeverity2.Warning,
            message: localize7("additionalItemsWarning", "Array has too many items according to schema. Expected {0} or fewer.", subSchemas.length),
            source: getSchemaSource(schema2, originalSchema),
            schemaUri: getSchemaUri(schema2, originalSchema)
          });
        }
      }
    } else {
      const itemSchema = asSchema2(schema2.items);
      if (itemSchema) {
        for (const item of node2.items) {
          const itemValidationResult = new ValidationResult2(isKubernetes);
          validate2(item, itemSchema, schema2, itemValidationResult, matchingSchemas2, options);
          validationResult2.mergePropertyMatch(itemValidationResult);
          validationResult2.mergeEnumValues(itemValidationResult);
        }
      }
    }
    const containsSchema = asSchema2(schema2.contains);
    if (containsSchema) {
      const doesContain = node2.items.some((item) => {
        const itemValidationResult = new ValidationResult2(isKubernetes);
        validate2(item, containsSchema, schema2, itemValidationResult, NoOpSchemaCollector2.instance, options);
        return !itemValidationResult.hasProblems();
      });
      if (!doesContain) {
        validationResult2.problems.push({
          location: { offset: node2.offset, length: node2.length },
          severity: DiagnosticSeverity2.Warning,
          message: schema2.errorMessage || localize7("requiredItemMissingWarning", "Array does not contain required item."),
          source: getSchemaSource(schema2, originalSchema),
          schemaUri: getSchemaUri(schema2, originalSchema)
        });
      }
    }
    if (isNumber2(schema2.minItems) && node2.items.length < schema2.minItems) {
      validationResult2.problems.push({
        location: { offset: node2.offset, length: node2.length },
        severity: DiagnosticSeverity2.Warning,
        message: localize7("minItemsWarning", "Array has too few items. Expected {0} or more.", schema2.minItems),
        source: getSchemaSource(schema2, originalSchema),
        schemaUri: getSchemaUri(schema2, originalSchema)
      });
    }
    if (isNumber2(schema2.maxItems) && node2.items.length > schema2.maxItems) {
      validationResult2.problems.push({
        location: { offset: node2.offset, length: node2.length },
        severity: DiagnosticSeverity2.Warning,
        message: localize7("maxItemsWarning", "Array has too many items. Expected {0} or fewer.", schema2.maxItems),
        source: getSchemaSource(schema2, originalSchema),
        schemaUri: getSchemaUri(schema2, originalSchema)
      });
    }
    if (schema2.uniqueItems === true) {
      const values = getNodeValue4(node2);
      const duplicates = values.some((value, index) => {
        return index !== values.lastIndexOf(value);
      });
      if (duplicates) {
        validationResult2.problems.push({
          location: { offset: node2.offset, length: node2.length },
          severity: DiagnosticSeverity2.Warning,
          message: localize7("uniqueItemsWarning", "Array has duplicate items."),
          source: getSchemaSource(schema2, originalSchema),
          schemaUri: getSchemaUri(schema2, originalSchema)
        });
      }
    }
  }
  function _validateObjectNode(node2, schema2, validationResult2, matchingSchemas2) {
    var _a;
    const seenKeys = Object.create(null);
    const unprocessedProperties = [];
    const unprocessedNodes = [...node2.properties];
    while (unprocessedNodes.length > 0) {
      const propertyNode = unprocessedNodes.pop();
      const key = propertyNode.keyNode.value;
      if (key === "<<" && propertyNode.valueNode) {
        switch (propertyNode.valueNode.type) {
          case "object": {
            unprocessedNodes.push(...propertyNode.valueNode["properties"]);
            break;
          }
          case "array": {
            propertyNode.valueNode["items"].forEach((sequenceNode) => {
              if (sequenceNode && isIterable(sequenceNode["properties"])) {
                unprocessedNodes.push(...sequenceNode["properties"]);
              }
            });
            break;
          }
          default: {
            break;
          }
        }
      } else {
        seenKeys[key] = propertyNode.valueNode;
        unprocessedProperties.push(key);
      }
    }
    if (Array.isArray(schema2.required)) {
      for (const propertyName of schema2.required) {
        if (!seenKeys[propertyName]) {
          const keyNode = node2.parent && node2.parent.type === "property" && node2.parent.keyNode;
          const location = keyNode ? { offset: keyNode.offset, length: keyNode.length } : { offset: node2.offset, length: 1 };
          validationResult2.problems.push({
            location,
            severity: DiagnosticSeverity2.Warning,
            message: getWarningMessage(ProblemType.missingRequiredPropWarning, [propertyName]),
            source: getSchemaSource(schema2, originalSchema),
            schemaUri: getSchemaUri(schema2, originalSchema),
            problemArgs: [propertyName],
            problemType: ProblemType.missingRequiredPropWarning
          });
        }
      }
    }
    const propertyProcessed = (prop) => {
      let index = unprocessedProperties.indexOf(prop);
      while (index >= 0) {
        unprocessedProperties.splice(index, 1);
        index = unprocessedProperties.indexOf(prop);
      }
    };
    if (schema2.properties) {
      for (const propertyName of Object.keys(schema2.properties)) {
        propertyProcessed(propertyName);
        const propertySchema = schema2.properties[propertyName];
        const child = seenKeys[propertyName];
        if (child) {
          if (isBoolean2(propertySchema)) {
            if (!propertySchema) {
              const propertyNode = child.parent;
              validationResult2.problems.push({
                location: {
                  offset: propertyNode.keyNode.offset,
                  length: propertyNode.keyNode.length
                },
                severity: DiagnosticSeverity2.Warning,
                message: schema2.errorMessage || localize7("DisallowedExtraPropWarning", "Property {0} is not allowed.", propertyName),
                source: getSchemaSource(schema2, originalSchema),
                schemaUri: getSchemaUri(schema2, originalSchema)
              });
            } else {
              validationResult2.propertiesMatches++;
              validationResult2.propertiesValueMatches++;
            }
          } else {
            propertySchema.url = (_a = schema2.url) !== null && _a !== void 0 ? _a : originalSchema.url;
            const propertyValidationResult = new ValidationResult2(isKubernetes);
            validate2(child, propertySchema, schema2, propertyValidationResult, matchingSchemas2, options);
            validationResult2.mergePropertyMatch(propertyValidationResult);
            validationResult2.mergeEnumValues(propertyValidationResult);
          }
        }
      }
    }
    if (schema2.patternProperties) {
      for (const propertyPattern of Object.keys(schema2.patternProperties)) {
        const regex = safeCreateUnicodeRegExp(propertyPattern);
        for (const propertyName of unprocessedProperties.slice(0)) {
          if (regex.test(propertyName)) {
            propertyProcessed(propertyName);
            const child = seenKeys[propertyName];
            if (child) {
              const propertySchema = schema2.patternProperties[propertyPattern];
              if (isBoolean2(propertySchema)) {
                if (!propertySchema) {
                  const propertyNode = child.parent;
                  validationResult2.problems.push({
                    location: {
                      offset: propertyNode.keyNode.offset,
                      length: propertyNode.keyNode.length
                    },
                    severity: DiagnosticSeverity2.Warning,
                    message: schema2.errorMessage || localize7("DisallowedExtraPropWarning", "Property {0} is not allowed.", propertyName),
                    source: getSchemaSource(schema2, originalSchema),
                    schemaUri: getSchemaUri(schema2, originalSchema)
                  });
                } else {
                  validationResult2.propertiesMatches++;
                  validationResult2.propertiesValueMatches++;
                }
              } else {
                const propertyValidationResult = new ValidationResult2(isKubernetes);
                validate2(child, propertySchema, schema2, propertyValidationResult, matchingSchemas2, options);
                validationResult2.mergePropertyMatch(propertyValidationResult);
                validationResult2.mergeEnumValues(propertyValidationResult);
              }
            }
          }
        }
      }
    }
    if (typeof schema2.additionalProperties === "object") {
      for (const propertyName of unprocessedProperties) {
        const child = seenKeys[propertyName];
        if (child) {
          const propertyValidationResult = new ValidationResult2(isKubernetes);
          validate2(child, schema2.additionalProperties, schema2, propertyValidationResult, matchingSchemas2, options);
          validationResult2.mergePropertyMatch(propertyValidationResult);
          validationResult2.mergeEnumValues(propertyValidationResult);
        }
      }
    } else if (schema2.additionalProperties === false || schema2.type === "object" && schema2.additionalProperties === void 0 && options.disableAdditionalProperties === true) {
      if (unprocessedProperties.length > 0) {
        for (const propertyName of unprocessedProperties) {
          const child = seenKeys[propertyName];
          if (child) {
            let propertyNode = null;
            if (child.type !== "property") {
              propertyNode = child.parent;
              if (propertyNode.type === "object") {
                propertyNode = propertyNode.properties[0];
              }
            } else {
              propertyNode = child;
            }
            validationResult2.problems.push({
              location: {
                offset: propertyNode.keyNode.offset,
                length: propertyNode.keyNode.length
              },
              severity: DiagnosticSeverity2.Warning,
              message: schema2.errorMessage || localize7("DisallowedExtraPropWarning", "Property {0} is not allowed.", propertyName),
              source: getSchemaSource(schema2, originalSchema),
              schemaUri: getSchemaUri(schema2, originalSchema)
            });
          }
        }
      }
    }
    if (isNumber2(schema2.maxProperties)) {
      if (node2.properties.length > schema2.maxProperties) {
        validationResult2.problems.push({
          location: { offset: node2.offset, length: node2.length },
          severity: DiagnosticSeverity2.Warning,
          message: localize7("MaxPropWarning", "Object has more properties than limit of {0}.", schema2.maxProperties),
          source: getSchemaSource(schema2, originalSchema),
          schemaUri: getSchemaUri(schema2, originalSchema)
        });
      }
    }
    if (isNumber2(schema2.minProperties)) {
      if (node2.properties.length < schema2.minProperties) {
        validationResult2.problems.push({
          location: { offset: node2.offset, length: node2.length },
          severity: DiagnosticSeverity2.Warning,
          message: localize7("MinPropWarning", "Object has fewer properties than the required number of {0}", schema2.minProperties),
          source: getSchemaSource(schema2, originalSchema),
          schemaUri: getSchemaUri(schema2, originalSchema)
        });
      }
    }
    if (schema2.dependencies) {
      for (const key of Object.keys(schema2.dependencies)) {
        const prop = seenKeys[key];
        if (prop) {
          const propertyDep = schema2.dependencies[key];
          if (Array.isArray(propertyDep)) {
            for (const requiredProp of propertyDep) {
              if (!seenKeys[requiredProp]) {
                validationResult2.problems.push({
                  location: { offset: node2.offset, length: node2.length },
                  severity: DiagnosticSeverity2.Warning,
                  message: localize7("RequiredDependentPropWarning", "Object is missing property {0} required by property {1}.", requiredProp, key),
                  source: getSchemaSource(schema2, originalSchema),
                  schemaUri: getSchemaUri(schema2, originalSchema)
                });
              } else {
                validationResult2.propertiesValueMatches++;
              }
            }
          } else {
            const propertySchema = asSchema2(propertyDep);
            if (propertySchema) {
              const propertyValidationResult = new ValidationResult2(isKubernetes);
              validate2(node2, propertySchema, schema2, propertyValidationResult, matchingSchemas2, options);
              validationResult2.mergePropertyMatch(propertyValidationResult);
              validationResult2.mergeEnumValues(propertyValidationResult);
            }
          }
        }
      }
    }
    const propertyNames = asSchema2(schema2.propertyNames);
    if (propertyNames) {
      for (const f2 of node2.properties) {
        const key = f2.keyNode;
        if (key) {
          validate2(key, propertyNames, schema2, validationResult2, NoOpSchemaCollector2.instance, options);
        }
      }
    }
  }
  function alternativeComparison(subValidationResult, bestMatch, subSchema, subMatchingSchemas) {
    const compareResult = subValidationResult.compareKubernetes(bestMatch.validationResult);
    if (compareResult > 0) {
      bestMatch = {
        schema: subSchema,
        validationResult: subValidationResult,
        matchingSchemas: subMatchingSchemas
      };
    } else if (compareResult === 0) {
      bestMatch.matchingSchemas.merge(subMatchingSchemas);
      bestMatch.validationResult.mergeEnumValues(subValidationResult);
    }
    return bestMatch;
  }
  function genericComparison(maxOneMatch, subValidationResult, bestMatch, subSchema, subMatchingSchemas) {
    if (!maxOneMatch && !subValidationResult.hasProblems() && !bestMatch.validationResult.hasProblems()) {
      bestMatch.matchingSchemas.merge(subMatchingSchemas);
      bestMatch.validationResult.propertiesMatches += subValidationResult.propertiesMatches;
      bestMatch.validationResult.propertiesValueMatches += subValidationResult.propertiesValueMatches;
    } else {
      const compareResult = subValidationResult.compareGeneric(bestMatch.validationResult);
      if (compareResult > 0) {
        bestMatch = {
          schema: subSchema,
          validationResult: subValidationResult,
          matchingSchemas: subMatchingSchemas
        };
      } else if (compareResult === 0) {
        bestMatch.matchingSchemas.merge(subMatchingSchemas);
        bestMatch.validationResult.mergeEnumValues(subValidationResult);
        bestMatch.validationResult.mergeWarningGeneric(subValidationResult, [
          ProblemType.missingRequiredPropWarning,
          ProblemType.typeMismatchWarning,
          ProblemType.constWarning
        ]);
      }
    }
    return bestMatch;
  }
}
function getSchemaSource(schema, originalSchema) {
  var _a;
  if (schema) {
    let label;
    if (schema.title) {
      label = schema.title;
    } else if (originalSchema.title) {
      label = originalSchema.title;
    } else {
      const uriString = (_a = schema.url) !== null && _a !== void 0 ? _a : originalSchema.url;
      if (uriString) {
        const url = URI.parse(uriString);
        if (url.scheme === "file") {
          label = url.fsPath;
        }
        label = url.toString();
      }
    }
    if (label) {
      return `${YAML_SCHEMA_PREFIX}${label}`;
    }
  }
  return YAML_SOURCE;
}
function getSchemaUri(schema, originalSchema) {
  var _a;
  const uriString = (_a = schema.url) !== null && _a !== void 0 ? _a : originalSchema.url;
  return uriString ? [uriString] : [];
}
function getWarningMessage(problemType, args) {
  return localize7(problemType, ProblemTypeMessages[problemType], args.join(" | "));
}

// node_modules/yaml-language-server/lib/esm/languageservice/parser/yaml-documents.js
import { isPair as isPair2, isScalar as isScalar3, visit as visit2 } from "yaml";

// node_modules/yaml-language-server/lib/esm/languageservice/parser/ast-converter.js
import { isScalar, isMap, isPair, isSeq, isNode, isAlias } from "yaml";
var maxRefCount = 1e3;
var refDepth = 0;
function convertAST(parent, node, doc, lineCounter) {
  if (!parent) {
    refDepth = 0;
  }
  if (!node) {
    return;
  }
  if (isMap(node)) {
    return convertMap(node, parent, doc, lineCounter);
  }
  if (isPair(node)) {
    return convertPair(node, parent, doc, lineCounter);
  }
  if (isSeq(node)) {
    return convertSeq(node, parent, doc, lineCounter);
  }
  if (isScalar(node)) {
    return convertScalar(node, parent);
  }
  if (isAlias(node)) {
    if (refDepth > maxRefCount) {
      return;
    }
    return convertAlias(node, parent, doc, lineCounter);
  }
}
function convertMap(node, parent, doc, lineCounter) {
  let range;
  if (node.flow && !node.range) {
    range = collectFlowMapRange(node);
  } else {
    range = node.range;
  }
  const result = new ObjectASTNodeImpl2(parent, node, ...toFixedOffsetLength(range, lineCounter));
  for (const it of node.items) {
    if (isPair(it)) {
      result.properties.push(convertAST(result, it, doc, lineCounter));
    }
  }
  return result;
}
function convertPair(node, parent, doc, lineCounter) {
  const keyNode = node.key;
  const valueNode = node.value;
  const rangeStart = keyNode.range[0];
  let rangeEnd = keyNode.range[1];
  let nodeEnd = keyNode.range[2];
  if (valueNode) {
    rangeEnd = valueNode.range[1];
    nodeEnd = valueNode.range[2];
  }
  const result = new PropertyASTNodeImpl2(parent, node, ...toFixedOffsetLength([rangeStart, rangeEnd, nodeEnd], lineCounter));
  if (isAlias(keyNode)) {
    const keyAlias = new StringASTNodeImpl2(parent, keyNode, ...toOffsetLength(keyNode.range));
    keyAlias.value = keyNode.source;
    result.keyNode = keyAlias;
  } else {
    result.keyNode = convertAST(result, keyNode, doc, lineCounter);
  }
  result.valueNode = convertAST(result, valueNode, doc, lineCounter);
  return result;
}
function convertSeq(node, parent, doc, lineCounter) {
  const result = new ArrayASTNodeImpl2(parent, node, ...toOffsetLength(node.range));
  for (const it of node.items) {
    if (isNode(it)) {
      result.children.push(convertAST(result, it, doc, lineCounter));
    }
  }
  return result;
}
function convertScalar(node, parent) {
  if (node.value === null) {
    return new NullASTNodeImpl2(parent, node, ...toOffsetLength(node.range));
  }
  switch (typeof node.value) {
    case "string": {
      const result = new StringASTNodeImpl2(parent, node, ...toOffsetLength(node.range));
      result.value = node.value;
      return result;
    }
    case "boolean":
      return new BooleanASTNodeImpl2(parent, node, node.value, ...toOffsetLength(node.range));
    case "number": {
      const result = new NumberASTNodeImpl2(parent, node, ...toOffsetLength(node.range));
      result.value = node.value;
      result.isInteger = Number.isInteger(result.value);
      return result;
    }
  }
}
function convertAlias(node, parent, doc, lineCounter) {
  refDepth++;
  return convertAST(parent, node.resolve(doc), doc, lineCounter);
}
function toOffsetLength(range) {
  return [range[0], range[1] - range[0]];
}
function toFixedOffsetLength(range, lineCounter) {
  const start = lineCounter.linePos(range[0]);
  const end = lineCounter.linePos(range[1]);
  const result = [range[0], range[1] - range[0]];
  if (start.line !== end.line && (lineCounter.lineStarts.length !== end.line || end.col === 1)) {
    result[1]--;
  }
  return result;
}
function collectFlowMapRange(node) {
  let start = Number.MAX_SAFE_INTEGER;
  let end = 0;
  for (const it of node.items) {
    if (isPair(it)) {
      if (isNode(it.key)) {
        if (it.key.range && it.key.range[0] <= start) {
          start = it.key.range[0];
        }
      }
      if (isNode(it.value)) {
        if (it.value.range && it.value.range[2] >= end) {
          end = it.value.range[2];
        }
      }
    }
  }
  return [start, end, end];
}

// node_modules/yaml-language-server/lib/esm/languageservice/utils/astUtils.js
import { isDocument, isScalar as isScalar2, visit } from "yaml";
function getParent(doc, nodeToFind) {
  let parentNode;
  visit(doc, (_, node, path5) => {
    if (node === nodeToFind) {
      parentNode = path5[path5.length - 1];
      return visit.BREAK;
    }
  });
  if (isDocument(parentNode)) {
    return void 0;
  }
  return parentNode;
}
function isMapContainsEmptyPair(map) {
  if (map.items.length > 1) {
    return false;
  }
  const pair = map.items[0];
  if (isScalar2(pair.key) && isScalar2(pair.value) && pair.key.value === "" && !pair.value.value) {
    return true;
  }
  return false;
}
function indexOf(seq, item) {
  for (const [i, obj] of seq.items.entries()) {
    if (item === obj) {
      return i;
    }
  }
  return void 0;
}
function isInComment(tokens, offset) {
  let inComment = false;
  for (const token of tokens) {
    if (token.type === "document") {
      _visit([], token, (item) => {
        var _a;
        if (isCollectionItem(item) && ((_a = item.value) === null || _a === void 0 ? void 0 : _a.type) === "comment") {
          if (token.offset <= offset && item.value.source.length + item.value.offset >= offset) {
            inComment = true;
            return visit.BREAK;
          }
        } else if (item.type === "comment" && item.offset <= offset && item.offset + item.source.length >= offset) {
          inComment = true;
          return visit.BREAK;
        }
      });
    } else if (token.type === "comment") {
      if (token.offset <= offset && token.source.length + token.offset >= offset) {
        return true;
      }
    }
    if (inComment) {
      break;
    }
  }
  return inComment;
}
function isCollectionItem(token) {
  return token["start"] !== void 0;
}
function _visit(path5, item, visitor) {
  let ctrl = visitor(item, path5);
  if (typeof ctrl === "symbol")
    return ctrl;
  for (const field of ["key", "value"]) {
    const token2 = item[field];
    if (token2 && "items" in token2) {
      for (let i = 0; i < token2.items.length; ++i) {
        const ci = _visit(Object.freeze(path5.concat([[field, i]])), token2.items[i], visitor);
        if (typeof ci === "number")
          i = ci - 1;
        else if (ci === visit.BREAK)
          return visit.BREAK;
        else if (ci === visit.REMOVE) {
          token2.items.splice(i, 1);
          i -= 1;
        }
      }
      if (typeof ctrl === "function" && field === "key")
        ctrl = ctrl(item, path5);
    }
  }
  const token = item["sep"];
  if (token) {
    for (let i = 0; i < token.length; ++i) {
      const ci = _visit(Object.freeze(path5), token[i], visitor);
      if (typeof ci === "number")
        i = ci - 1;
      else if (ci === visit.BREAK)
        return visit.BREAK;
      else if (ci === visit.REMOVE) {
        token.items.splice(i, 1);
        i -= 1;
      }
    }
  }
  return typeof ctrl === "function" ? ctrl(item, path5) : ctrl;
}

// node_modules/yaml-language-server/lib/esm/languageservice/parser/yaml-documents.js
var SingleYAMLDocument = class extends JSONDocument2 {
  constructor(lineCounter) {
    super(null, []);
    this.lineCounter = lineCounter;
  }
  collectLineComments() {
    this._lineComments = [];
    if (this._internalDocument.commentBefore) {
      this._lineComments.push(`#${this._internalDocument.commentBefore}`);
    }
    visit2(this.internalDocument, (_key, node) => {
      if (node === null || node === void 0 ? void 0 : node.commentBefore) {
        this._lineComments.push(`#${node.commentBefore}`);
      }
      if (node === null || node === void 0 ? void 0 : node.comment) {
        this._lineComments.push(`#${node.comment}`);
      }
    });
    if (this._internalDocument.comment) {
      this._lineComments.push(`#${this._internalDocument.comment}`);
    }
  }
  set internalDocument(document) {
    this._internalDocument = document;
    this.root = convertAST(null, this._internalDocument.contents, this._internalDocument, this.lineCounter);
  }
  get internalDocument() {
    return this._internalDocument;
  }
  get lineComments() {
    if (!this._lineComments) {
      this.collectLineComments();
    }
    return this._lineComments;
  }
  set lineComments(val) {
    this._lineComments = val;
  }
  get errors() {
    return this.internalDocument.errors.map(YAMLErrorToYamlDocDiagnostics);
  }
  get warnings() {
    return this.internalDocument.warnings.map(YAMLErrorToYamlDocDiagnostics);
  }
  getSchemas(schema, doc, node) {
    const matchingSchemas = [];
    doc.validate(schema, matchingSchemas, node.start);
    return matchingSchemas;
  }
  getNodeFromPosition(positionOffset, textBuffer) {
    const position = textBuffer.getPosition(positionOffset);
    const lineContent = textBuffer.getLineContent(position.line);
    if (lineContent.trim().length === 0) {
      return [this.findClosestNode(positionOffset, textBuffer), true];
    }
    let closestNode;
    visit2(this.internalDocument, (key, node) => {
      if (!node) {
        return;
      }
      const range = node.range;
      if (!range) {
        return;
      }
      if (range[0] <= positionOffset && range[1] >= positionOffset) {
        closestNode = node;
      } else {
        return visit2.SKIP;
      }
    });
    return [closestNode, false];
  }
  findClosestNode(offset, textBuffer) {
    let offsetDiff = this.internalDocument.range[2];
    let maxOffset = this.internalDocument.range[0];
    let closestNode;
    visit2(this.internalDocument, (key, node) => {
      if (!node) {
        return;
      }
      const range = node.range;
      if (!range) {
        return;
      }
      const diff = Math.abs(range[2] - offset);
      if (maxOffset <= range[0] && diff <= offsetDiff) {
        offsetDiff = diff;
        maxOffset = range[0];
        closestNode = node;
      }
    });
    const position = textBuffer.getPosition(offset);
    const lineContent = textBuffer.getLineContent(position.line);
    const indentation = getIndentation(lineContent, position.character);
    if (isScalar3(closestNode) && closestNode.value === null) {
      return closestNode;
    }
    if (indentation === position.character) {
      closestNode = this.getProperParentByIndentation(indentation, closestNode, textBuffer);
    }
    return closestNode;
  }
  getProperParentByIndentation(indentation, node, textBuffer) {
    if (!node) {
      return this.internalDocument.contents;
    }
    if (node.range) {
      const position = textBuffer.getPosition(node.range[0]);
      if (position.character !== indentation && position.character > 0) {
        const parent = this.getParent(node);
        if (parent) {
          return this.getProperParentByIndentation(indentation, parent, textBuffer);
        }
      } else {
        return node;
      }
    } else if (isPair2(node)) {
      const parent = this.getParent(node);
      return this.getProperParentByIndentation(indentation, parent, textBuffer);
    }
    return node;
  }
  getParent(node) {
    return getParent(this.internalDocument, node);
  }
};
var YAMLDocument = class {
  constructor(documents, tokens) {
    this.documents = documents;
    this.tokens = tokens;
    this.errors = [];
    this.warnings = [];
  }
};
var YamlDocuments = class {
  constructor() {
    this.cache = new Map();
  }
  getYamlDocument(document, parserOptions, addRootObject = false) {
    this.ensureCache(document, parserOptions !== null && parserOptions !== void 0 ? parserOptions : defaultOptions, addRootObject);
    return this.cache.get(document.uri).document;
  }
  clear() {
    this.cache.clear();
  }
  ensureCache(document, parserOptions, addRootObject) {
    const key = document.uri;
    if (!this.cache.has(key)) {
      this.cache.set(key, { version: -1, document: new YAMLDocument([], []), parserOptions: defaultOptions });
    }
    const cacheEntry = this.cache.get(key);
    if (cacheEntry.version !== document.version || parserOptions.customTags && !isArrayEqual(cacheEntry.parserOptions.customTags, parserOptions.customTags)) {
      let text = document.getText();
      if (addRootObject && !/\S/.test(text)) {
        text = `{${text}}`;
      }
      const doc = parse3(text, parserOptions);
      cacheEntry.document = doc;
      cacheEntry.version = document.version;
      cacheEntry.parserOptions = parserOptions;
    }
  }
};
var yamlDocumentsCache = new YamlDocuments();
function YAMLErrorToYamlDocDiagnostics(error) {
  return {
    message: error.message,
    location: {
      start: error.pos[0],
      end: error.pos[1],
      toLineEnd: true
    },
    severity: 1,
    code: ErrorCode.Undefined
  };
}

// node_modules/yaml-language-server/lib/esm/languageservice/parser/custom-tag-provider.js
import { isSeq as isSeq2, isMap as isMap2 } from "yaml";
var CommonTagImpl = class {
  constructor(tag, type) {
    this.tag = tag;
    this.type = type;
  }
  get collection() {
    if (this.type === "mapping") {
      return "map";
    }
    if (this.type === "sequence") {
      return "seq";
    }
    return void 0;
  }
  resolve(value) {
    if (isMap2(value) && this.type === "mapping") {
      return value;
    }
    if (isSeq2(value) && this.type === "sequence") {
      return value;
    }
    if (typeof value === "string" && this.type === "scalar") {
      return value;
    }
  }
};
var IncludeTag = class {
  constructor() {
    this.tag = "!include";
    this.type = "scalar";
  }
  resolve(value, onError) {
    if (value && value.length > 0 && value.trim()) {
      return value;
    }
    onError("!include without value");
  }
};
function getCustomTags(customTags) {
  const tags = [];
  const filteredTags = filterInvalidCustomTags(customTags);
  for (const tag of filteredTags) {
    const typeInfo = tag.split(" ");
    const tagName = typeInfo[0];
    const tagType = typeInfo[1] && typeInfo[1].toLowerCase() || "scalar";
    tags.push(new CommonTagImpl(tagName, tagType));
  }
  tags.push(new IncludeTag());
  return tags;
}

// node_modules/yaml-language-server/lib/esm/languageservice/parser/yamlParser07.js
"use strict";
var defaultOptions = {
  customTags: [],
  yamlVersion: "1.2"
};
function parse3(text, parserOptions = defaultOptions) {
  const options = {
    strict: false,
    customTags: getCustomTags(parserOptions.customTags),
    version: parserOptions.yamlVersion
  };
  const composer = new Composer(options);
  const lineCounter = new LineCounter();
  const parser2 = new Parser5(lineCounter.addNewLine);
  const tokens = parser2.parse(text);
  const tokensArr = Array.from(tokens);
  const docs = composer.compose(tokensArr, true, text.length);
  const yamlDocs = Array.from(docs, (doc) => parsedDocToSingleYAMLDocument(doc, lineCounter));
  return new YAMLDocument(yamlDocs, tokensArr);
}
function parsedDocToSingleYAMLDocument(parsedDoc, lineCounter) {
  const syd = new SingleYAMLDocument(lineCounter);
  syd.internalDocument = parsedDoc;
  return syd;
}

// node_modules/yaml-language-server/lib/esm/languageservice/services/modelineUtil.js
"use strict";
function getSchemaFromModeline(doc) {
  if (doc instanceof SingleYAMLDocument) {
    const yamlLanguageServerModeline = doc.lineComments.find((lineComment) => {
      return isModeline(lineComment);
    });
    if (yamlLanguageServerModeline != void 0) {
      const schemaMatchs = yamlLanguageServerModeline.match(/\$schema=\S+/g);
      if (schemaMatchs !== null && schemaMatchs.length >= 1) {
        if (schemaMatchs.length >= 2) {
          console.log("Several $schema attributes have been found on the yaml-language-server modeline. The first one will be picked.");
        }
        return schemaMatchs[0].substring("$schema=".length);
      }
    }
  }
  return void 0;
}
function isModeline(lineText) {
  const matchModeline = lineText.match(/^#\s+yaml-language-server\s*:/g);
  return matchModeline !== null && matchModeline.length === 1;
}

// node_modules/yaml-language-server/lib/esm/languageservice/services/yamlSchemaService.js
"use strict";
var __awaiter = function(thisArg, _arguments, P, generator) {
  function adopt(value) {
    return value instanceof P ? value : new P(function(resolve2) {
      resolve2(value);
    });
  }
  return new (P || (P = Promise))(function(resolve2, reject) {
    function fulfilled(value) {
      try {
        step(generator.next(value));
      } catch (e) {
        reject(e);
      }
    }
    function rejected(value) {
      try {
        step(generator["throw"](value));
      } catch (e) {
        reject(e);
      }
    }
    function step(result) {
      result.done ? resolve2(result.value) : adopt(result.value).then(fulfilled, rejected);
    }
    step((generator = generator.apply(thisArg, _arguments || [])).next());
  });
};
var localize8 = loadMessageBundle();
var MODIFICATION_ACTIONS;
(function(MODIFICATION_ACTIONS2) {
  MODIFICATION_ACTIONS2[MODIFICATION_ACTIONS2["delete"] = 0] = "delete";
  MODIFICATION_ACTIONS2[MODIFICATION_ACTIONS2["add"] = 1] = "add";
  MODIFICATION_ACTIONS2[MODIFICATION_ACTIONS2["deleteAll"] = 2] = "deleteAll";
})(MODIFICATION_ACTIONS || (MODIFICATION_ACTIONS = {}));
var YAMLSchemaService = class extends JSONSchemaService {
  constructor(requestService, contextService, promiseConstructor) {
    super(requestService, contextService, promiseConstructor);
    this.schemaUriToNameAndDescription = new Map();
    this.customSchemaProvider = void 0;
    this.requestService = requestService;
    this.schemaPriorityMapping = new Map();
  }
  registerCustomSchemaProvider(customSchemaProvider) {
    this.customSchemaProvider = customSchemaProvider;
  }
  getAllSchemas() {
    const result = [];
    const schemaUris = new Set();
    for (const filePattern of this.filePatternAssociations) {
      const schemaUri = filePattern.uris[0];
      if (schemaUris.has(schemaUri)) {
        continue;
      }
      schemaUris.add(schemaUri);
      const schemaHandle = {
        uri: schemaUri,
        fromStore: false,
        usedForCurrentFile: false
      };
      if (this.schemaUriToNameAndDescription.has(schemaUri)) {
        const [name, description] = this.schemaUriToNameAndDescription.get(schemaUri);
        schemaHandle.name = name;
        schemaHandle.description = description;
        schemaHandle.fromStore = true;
      }
      result.push(schemaHandle);
    }
    return result;
  }
  resolveSchemaContent(schemaToResolve, schemaURL, dependencies) {
    return __awaiter(this, void 0, void 0, function* () {
      const resolveErrors = schemaToResolve.errors.slice(0);
      let schema = schemaToResolve.schema;
      const contextService = this.contextService;
      const findSection = (schema2, path5) => {
        if (!path5) {
          return schema2;
        }
        let current = schema2;
        if (path5[0] === "/") {
          path5 = path5.substr(1);
        }
        path5.split("/").some((part) => {
          current = current[part];
          return !current;
        });
        return current;
      };
      const merge = (target, sourceRoot, sourceURI, path5) => {
        const section = findSection(sourceRoot, path5);
        if (section) {
          for (const key in section) {
            if (Object.prototype.hasOwnProperty.call(section, key) && !Object.prototype.hasOwnProperty.call(target, key)) {
              target[key] = section[key];
            }
          }
        } else {
          resolveErrors.push(localize8("json.schema.invalidref", "$ref '{0}' in '{1}' can not be resolved.", path5, sourceURI));
        }
      };
      const resolveExternalLink = (node, uri, linkPath, parentSchemaURL, parentSchemaDependencies) => {
        if (contextService && !/^\w+:\/\/.*/.test(uri)) {
          uri = contextService.resolveRelativePath(uri, parentSchemaURL);
        }
        uri = this.normalizeId(uri);
        const referencedHandle = this.getOrAddSchemaHandle(uri);
        return referencedHandle.getUnresolvedSchema().then((unresolvedSchema) => {
          parentSchemaDependencies[uri] = true;
          if (unresolvedSchema.errors.length) {
            const loc = linkPath ? uri + "#" + linkPath : uri;
            resolveErrors.push(localize8("json.schema.problemloadingref", "Problems loading reference '{0}': {1}", loc, unresolvedSchema.errors[0]));
          }
          merge(node, unresolvedSchema.schema, uri, linkPath);
          node.url = uri;
          return resolveRefs(node, unresolvedSchema.schema, uri, referencedHandle.dependencies);
        });
      };
      const resolveRefs = (node, parentSchema, parentSchemaURL, parentSchemaDependencies) => __awaiter(this, void 0, void 0, function* () {
        if (!node || typeof node !== "object") {
          return null;
        }
        const toWalk = [node];
        const seen = [];
        const openPromises = [];
        const collectEntries = (...entries) => {
          for (const entry of entries) {
            if (typeof entry === "object") {
              toWalk.push(entry);
            }
          }
        };
        const collectMapEntries = (...maps) => {
          for (const map of maps) {
            if (typeof map === "object") {
              for (const key in map) {
                const entry = map[key];
                if (typeof entry === "object") {
                  toWalk.push(entry);
                }
              }
            }
          }
        };
        const collectArrayEntries = (...arrays) => {
          for (const array of arrays) {
            if (Array.isArray(array)) {
              for (const entry of array) {
                if (typeof entry === "object") {
                  toWalk.push(entry);
                }
              }
            }
          }
        };
        const handleRef = (next) => {
          const seenRefs = [];
          while (next.$ref) {
            const ref = next.$ref;
            const segments = ref.split("#", 2);
            next._$ref = next.$ref;
            delete next.$ref;
            if (segments[0].length > 0) {
              openPromises.push(resolveExternalLink(next, segments[0], segments[1], parentSchemaURL, parentSchemaDependencies));
              return;
            } else {
              if (seenRefs.indexOf(ref) === -1) {
                merge(next, parentSchema, parentSchemaURL, segments[1]);
                seenRefs.push(ref);
              }
            }
          }
          collectEntries(next.items, next.additionalItems, next.additionalProperties, next.not, next.contains, next.propertyNames, next.if, next.then, next.else);
          collectMapEntries(next.definitions, next.properties, next.patternProperties, next.dependencies);
          collectArrayEntries(next.anyOf, next.allOf, next.oneOf, next.items, next.schemaSequence);
        };
        if (parentSchemaURL.indexOf("#") > 0) {
          const segments = parentSchemaURL.split("#", 2);
          if (segments[0].length > 0 && segments[1].length > 0) {
            const newSchema = {};
            yield resolveExternalLink(newSchema, segments[0], segments[1], parentSchemaURL, parentSchemaDependencies);
            for (const key in schema) {
              if (key === "required") {
                continue;
              }
              if (Object.prototype.hasOwnProperty.call(schema, key) && !Object.prototype.hasOwnProperty.call(newSchema, key)) {
                newSchema[key] = schema[key];
              }
            }
            schema = newSchema;
          }
        }
        while (toWalk.length) {
          const next = toWalk.pop();
          if (seen.indexOf(next) >= 0) {
            continue;
          }
          seen.push(next);
          handleRef(next);
        }
        return Promise.all(openPromises);
      });
      yield resolveRefs(schema, schema, schemaURL, dependencies);
      return new ResolvedSchema(schema, resolveErrors);
    });
  }
  getSchemaForResource(resource, doc) {
    const resolveSchema = () => {
      const seen = Object.create(null);
      const schemas = [];
      let schemaFromModeline = getSchemaFromModeline(doc);
      if (schemaFromModeline !== void 0) {
        if (!schemaFromModeline.startsWith("file:") && !schemaFromModeline.startsWith("http")) {
          if (!isAbsolute(schemaFromModeline)) {
            const resUri = URI.parse(resource);
            schemaFromModeline = URI.file(resolve(parse5(resUri.fsPath).dir, schemaFromModeline)).toString();
          } else {
            schemaFromModeline = URI.file(schemaFromModeline).toString();
          }
        }
        this.addSchemaPriority(schemaFromModeline, SchemaPriority.Modeline);
        schemas.push(schemaFromModeline);
        seen[schemaFromModeline] = true;
      }
      for (const entry of this.filePatternAssociations) {
        if (entry.matchesPattern(resource)) {
          for (const schemaId of entry.getURIs()) {
            if (!seen[schemaId]) {
              schemas.push(schemaId);
              seen[schemaId] = true;
            }
          }
        }
      }
      const normalizedResourceID = this.normalizeId(resource);
      if (this.schemasById[normalizedResourceID]) {
        schemas.push(normalizedResourceID);
      }
      if (schemas.length > 0) {
        const highestPrioSchemas = this.highestPrioritySchemas(schemas);
        const schemaHandle = super.createCombinedSchema(resource, highestPrioSchemas);
        return schemaHandle.getResolvedSchema().then((schema) => {
          if (schema.schema && typeof schema.schema !== "string") {
            schema.schema.url = schemaHandle.url;
          }
          if (schema.schema && schema.schema.schemaSequence && schema.schema.schemaSequence[doc.currentDocIndex]) {
            return new ResolvedSchema(schema.schema.schemaSequence[doc.currentDocIndex]);
          }
          return schema;
        });
      }
      return Promise.resolve(null);
    };
    if (this.customSchemaProvider) {
      return this.customSchemaProvider(resource).then((schemaUri) => {
        if (Array.isArray(schemaUri)) {
          if (schemaUri.length === 0) {
            return resolveSchema();
          }
          return Promise.all(schemaUri.map((schemaUri2) => {
            return this.resolveCustomSchema(schemaUri2, doc);
          })).then((schemas) => {
            return {
              errors: [],
              schema: {
                anyOf: schemas.map((schemaObj) => {
                  return schemaObj.schema;
                })
              }
            };
          }, () => {
            return resolveSchema();
          });
        }
        if (!schemaUri) {
          return resolveSchema();
        }
        return this.resolveCustomSchema(schemaUri, doc);
      }).then((schema) => {
        return schema;
      }, () => {
        return resolveSchema();
      });
    } else {
      return resolveSchema();
    }
  }
  addSchemaPriority(uri, priority) {
    let currSchemaArray = this.schemaPriorityMapping.get(uri);
    if (currSchemaArray) {
      currSchemaArray = currSchemaArray.add(priority);
      this.schemaPriorityMapping.set(uri, currSchemaArray);
    } else {
      this.schemaPriorityMapping.set(uri, new Set().add(priority));
    }
  }
  highestPrioritySchemas(schemas) {
    let highestPrio = 0;
    const priorityMapping = new Map();
    schemas.forEach((schema) => {
      const priority = this.schemaPriorityMapping.get(schema) || [0];
      priority.forEach((prio) => {
        if (prio > highestPrio) {
          highestPrio = prio;
        }
        let currPriorityArray = priorityMapping.get(prio);
        if (currPriorityArray) {
          currPriorityArray = currPriorityArray.concat(schema);
          priorityMapping.set(prio, currPriorityArray);
        } else {
          priorityMapping.set(prio, [schema]);
        }
      });
    });
    return priorityMapping.get(highestPrio) || [];
  }
  resolveCustomSchema(schemaUri, doc) {
    return __awaiter(this, void 0, void 0, function* () {
      const unresolvedSchema = yield this.loadSchema(schemaUri);
      const schema = yield this.resolveSchemaContent(unresolvedSchema, schemaUri, []);
      if (schema.schema) {
        schema.schema.url = schemaUri;
      }
      if (schema.schema && schema.schema.schemaSequence && schema.schema.schemaSequence[doc.currentDocIndex]) {
        return new ResolvedSchema(schema.schema.schemaSequence[doc.currentDocIndex]);
      }
      return schema;
    });
  }
  saveSchema(schemaId, schemaContent) {
    return __awaiter(this, void 0, void 0, function* () {
      const id = this.normalizeId(schemaId);
      this.getOrAddSchemaHandle(id, schemaContent);
      this.schemaPriorityMapping.set(id, new Set().add(SchemaPriority.Settings));
      return Promise.resolve(void 0);
    });
  }
  deleteSchemas(deletions) {
    return __awaiter(this, void 0, void 0, function* () {
      deletions.schemas.forEach((s) => {
        this.deleteSchema(s);
      });
      return Promise.resolve(void 0);
    });
  }
  deleteSchema(schemaId) {
    return __awaiter(this, void 0, void 0, function* () {
      const id = this.normalizeId(schemaId);
      if (this.schemasById[id]) {
        delete this.schemasById[id];
      }
      this.schemaPriorityMapping.delete(id);
      return Promise.resolve(void 0);
    });
  }
  addContent(additions) {
    return __awaiter(this, void 0, void 0, function* () {
      const schema = yield this.getResolvedSchema(additions.schema);
      if (schema) {
        const resolvedSchemaLocation = this.resolveJSONSchemaToSection(schema.schema, additions.path);
        if (typeof resolvedSchemaLocation === "object") {
          resolvedSchemaLocation[additions.key] = additions.content;
        }
        yield this.saveSchema(additions.schema, schema.schema);
      }
    });
  }
  deleteContent(deletions) {
    return __awaiter(this, void 0, void 0, function* () {
      const schema = yield this.getResolvedSchema(deletions.schema);
      if (schema) {
        const resolvedSchemaLocation = this.resolveJSONSchemaToSection(schema.schema, deletions.path);
        if (typeof resolvedSchemaLocation === "object") {
          delete resolvedSchemaLocation[deletions.key];
        }
        yield this.saveSchema(deletions.schema, schema.schema);
      }
    });
  }
  resolveJSONSchemaToSection(schema, paths) {
    const splitPathway = paths.split("/");
    let resolvedSchemaLocation = schema;
    for (const path5 of splitPathway) {
      if (path5 === "") {
        continue;
      }
      this.resolveNext(resolvedSchemaLocation, path5);
      resolvedSchemaLocation = resolvedSchemaLocation[path5];
    }
    return resolvedSchemaLocation;
  }
  resolveNext(object, token) {
    if (Array.isArray(object) && isNaN(token)) {
      throw new Error("Expected a number after the array object");
    } else if (typeof object === "object" && typeof token !== "string") {
      throw new Error("Expected a string after the object");
    }
  }
  normalizeId(id) {
    try {
      return URI.parse(id).toString();
    } catch (e) {
      return id;
    }
  }
  getOrAddSchemaHandle(id, unresolvedSchemaContent) {
    return super.getOrAddSchemaHandle(id, unresolvedSchemaContent);
  }
  loadSchema(schemaUri) {
    const requestService = this.requestService;
    return super.loadSchema(schemaUri).then((unresolvedJsonSchema) => {
      if (unresolvedJsonSchema.errors && unresolvedJsonSchema.schema === void 0) {
        return requestService(schemaUri).then((content) => {
          if (!content) {
            const errorMessage = localize8("json.schema.nocontent", "Unable to load schema from '{0}': No content.", toDisplayString2(schemaUri));
            return new UnresolvedSchema({}, [errorMessage]);
          }
          try {
            const schemaContent = parse4(content);
            return new UnresolvedSchema(schemaContent, []);
          } catch (yamlError) {
            const errorMessage = localize8("json.schema.invalidFormat", "Unable to parse content from '{0}': {1}.", toDisplayString2(schemaUri), yamlError);
            return new UnresolvedSchema({}, [errorMessage]);
          }
        }, (error) => {
          let errorMessage = error.toString();
          const errorSplit = error.toString().split("Error: ");
          if (errorSplit.length > 1) {
            errorMessage = errorSplit[1];
          }
          return new UnresolvedSchema({}, [errorMessage]);
        });
      }
      unresolvedJsonSchema.uri = schemaUri;
      if (this.schemaUriToNameAndDescription.has(schemaUri)) {
        const [name, description] = this.schemaUriToNameAndDescription.get(schemaUri);
        unresolvedJsonSchema.schema.title = name !== null && name !== void 0 ? name : unresolvedJsonSchema.schema.title;
        unresolvedJsonSchema.schema.description = description !== null && description !== void 0 ? description : unresolvedJsonSchema.schema.description;
      }
      return unresolvedJsonSchema;
    });
  }
  registerExternalSchema(uri, filePatterns, unresolvedSchema, name, description) {
    if (name || description) {
      this.schemaUriToNameAndDescription.set(uri, [name, description]);
    }
    return super.registerExternalSchema(uri, filePatterns, unresolvedSchema);
  }
  clearExternalSchemas() {
    super.clearExternalSchemas();
  }
  setSchemaContributions(schemaContributions2) {
    super.setSchemaContributions(schemaContributions2);
  }
  getRegisteredSchemaIds(filter) {
    return super.getRegisteredSchemaIds(filter);
  }
  getResolvedSchema(schemaId) {
    return super.getResolvedSchema(schemaId);
  }
  onResourceChange(uri) {
    return super.onResourceChange(uri);
  }
};
function toDisplayString2(url) {
  try {
    const uri = URI.parse(url);
    if (uri.scheme === "file") {
      return uri.fsPath;
    }
  } catch (e) {
  }
  return url;
}

// node_modules/yaml-language-server/lib/esm/languageservice/services/documentSymbols.js
"use strict";
var YAMLDocumentSymbols = class {
  constructor(schemaService, telemetry) {
    this.telemetry = telemetry;
    this.jsonDocumentSymbols = new JSONDocumentSymbols(schemaService);
    const origKeyLabel = this.jsonDocumentSymbols.getKeyLabel;
    this.jsonDocumentSymbols.getKeyLabel = (property) => {
      if (typeof property.keyNode.value === "object") {
        return property.keyNode.value.value;
      } else {
        return origKeyLabel.call(this.jsonDocumentSymbols, property);
      }
    };
  }
  findDocumentSymbols(document, context = { resultLimit: Number.MAX_VALUE }) {
    let results = [];
    try {
      const doc = yamlDocumentsCache.getYamlDocument(document);
      if (!doc || doc["documents"].length === 0) {
        return null;
      }
      for (const yamlDoc of doc["documents"]) {
        if (yamlDoc.root) {
          results = results.concat(this.jsonDocumentSymbols.findDocumentSymbols(document, yamlDoc, context));
        }
      }
    } catch (err) {
      this.telemetry.sendError("yaml.documentSymbols.error", { error: err, documentUri: document.uri });
    }
    return results;
  }
  findHierarchicalDocumentSymbols(document, context = { resultLimit: Number.MAX_VALUE }) {
    let results = [];
    try {
      const doc = yamlDocumentsCache.getYamlDocument(document);
      if (!doc || doc["documents"].length === 0) {
        return null;
      }
      for (const yamlDoc of doc["documents"]) {
        if (yamlDoc.root) {
          results = results.concat(this.jsonDocumentSymbols.findDocumentSymbols2(document, yamlDoc, context));
        }
      }
    } catch (err) {
      this.telemetry.sendError("yaml.hierarchicalDocumentSymbols.error", { error: err, documentUri: document.uri });
    }
    return results;
  }
};

// node_modules/yaml-language-server/lib/esm/languageservice/services/yamlHover.js
import { Range as Range3 } from "vscode-languageserver-types";

// node_modules/yaml-language-server/lib/esm/languageservice/parser/isKubernetes.js
function setKubernetesParserOption(jsonDocuments, option) {
  for (const jsonDoc of jsonDocuments) {
    jsonDoc.isKubernetes = option;
  }
}

// node_modules/yaml-language-server/lib/esm/languageservice/services/yamlHover.js
import {
  basename
} from "path-browserify";
"use strict";
var YAMLHover = class {
  constructor(schemaService, telemetry) {
    this.telemetry = telemetry;
    this.shouldHover = true;
    this.schemaService = schemaService;
  }
  configure(languageSettings) {
    if (languageSettings) {
      this.shouldHover = languageSettings.hover;
    }
  }
  doHover(document, position, isKubernetes = false) {
    try {
      if (!this.shouldHover || !document) {
        return Promise.resolve(void 0);
      }
      const doc = yamlDocumentsCache.getYamlDocument(document);
      const offset = document.offsetAt(position);
      const currentDoc = matchOffsetToDocument(offset, doc);
      if (currentDoc === null) {
        return Promise.resolve(void 0);
      }
      setKubernetesParserOption(doc.documents, isKubernetes);
      const currentDocIndex = doc.documents.indexOf(currentDoc);
      currentDoc.currentDocIndex = currentDocIndex;
      return this.getHover(document, position, currentDoc);
    } catch (error) {
      this.telemetry.sendError("yaml.hover.error", { error, documentUri: document.uri });
    }
  }
  getHover(document, position, doc) {
    const offset = document.offsetAt(position);
    let node = doc.getNodeFromOffset(offset);
    if (!node || (node.type === "object" || node.type === "array") && offset > node.offset + 1 && offset < node.offset + node.length - 1) {
      return Promise.resolve(null);
    }
    const hoverRangeNode = node;
    if (node.type === "string") {
      const parent = node.parent;
      if (parent && parent.type === "property" && parent.keyNode === node) {
        node = parent.valueNode;
        if (!node) {
          return Promise.resolve(null);
        }
      }
    }
    const hoverRange = Range3.create(document.positionAt(hoverRangeNode.offset), document.positionAt(hoverRangeNode.offset + hoverRangeNode.length));
    const createHover = (contents) => {
      const markupContent = {
        kind: "markdown",
        value: contents
      };
      const result = {
        contents: markupContent,
        range: hoverRange
      };
      return result;
    };
    return this.schemaService.getSchemaForResource(document.uri, doc).then((schema) => {
      if (schema && node && !schema.errors.length) {
        const matchingSchemas = doc.getMatchingSchemas(schema.schema, node.offset);
        let title = void 0;
        let markdownDescription = void 0;
        let markdownEnumValueDescription = void 0, enumValue = void 0;
        matchingSchemas.every((s) => {
          if (s.node === node && !s.inverted && s.schema) {
            title = title || s.schema.title;
            markdownDescription = markdownDescription || s.schema.markdownDescription || toMarkdown2(s.schema.description);
            if (s.schema.enum) {
              const idx = s.schema.enum.indexOf(getNodeValue4(node));
              if (s.schema.markdownEnumDescriptions) {
                markdownEnumValueDescription = s.schema.markdownEnumDescriptions[idx];
              } else if (s.schema.enumDescriptions) {
                markdownEnumValueDescription = toMarkdown2(s.schema.enumDescriptions[idx]);
              }
              if (markdownEnumValueDescription) {
                enumValue = s.schema.enum[idx];
                if (typeof enumValue !== "string") {
                  enumValue = JSON.stringify(enumValue);
                }
              }
            }
          }
          return true;
        });
        let result = "";
        if (title) {
          result = "#### " + toMarkdown2(title);
        }
        if (markdownDescription) {
          if (result.length > 0) {
            result += "\n\n";
          }
          result += markdownDescription;
        }
        if (markdownEnumValueDescription) {
          if (result.length > 0) {
            result += "\n\n";
          }
          result += `\`${toMarkdownCodeBlock2(enumValue)}\`: ${markdownEnumValueDescription}`;
        }
        if (result.length > 0 && schema.schema.url) {
          result += `

Source: [${getSchemaName(schema.schema)}](${schema.schema.url})`;
        }
        return createHover(result);
      }
      return null;
    });
  }
};
function getSchemaName(schema) {
  let result = "JSON Schema";
  const urlString = schema.url;
  if (urlString) {
    const url = URI.parse(urlString);
    result = basename(url.fsPath);
  } else if (schema.title) {
    result = schema.title;
  }
  return result;
}
function toMarkdown2(plain) {
  if (plain) {
    const res = plain.replace(/([^\n\r])(\r?\n)([^\n\r])/gm, "$1\n\n$3");
    return res.replace(/[\\`*_{}[\]()#+\-.!]/g, "\\$&");
  }
  return void 0;
}
function toMarkdownCodeBlock2(content) {
  if (content.indexOf("`") !== -1) {
    return "`` " + content + " ``";
  }
  return content;
}

// node_modules/yaml-language-server/lib/esm/languageservice/services/yamlValidation.js
import { Diagnostic as Diagnostic3, Position as Position2 } from "vscode-languageserver-types";

// node_modules/yaml-language-server/lib/esm/languageservice/utils/textBuffer.js
import { Range as Range4 } from "vscode-languageserver-types";
var TextBuffer = class {
  constructor(doc) {
    this.doc = doc;
  }
  getLineCount() {
    return this.doc.lineCount;
  }
  getLineLength(lineNumber) {
    const lineOffsets = this.doc.getLineOffsets();
    if (lineNumber >= lineOffsets.length) {
      return this.doc.getText().length;
    } else if (lineNumber < 0) {
      return 0;
    }
    const nextLineOffset = lineNumber + 1 < lineOffsets.length ? lineOffsets[lineNumber + 1] : this.doc.getText().length;
    return nextLineOffset - lineOffsets[lineNumber];
  }
  getLineContent(lineNumber) {
    const lineOffsets = this.doc.getLineOffsets();
    if (lineNumber >= lineOffsets.length) {
      return this.doc.getText();
    } else if (lineNumber < 0) {
      return "";
    }
    const nextLineOffset = lineNumber + 1 < lineOffsets.length ? lineOffsets[lineNumber + 1] : this.doc.getText().length;
    return this.doc.getText().substring(lineOffsets[lineNumber], nextLineOffset);
  }
  getLineCharCode(lineNumber, index) {
    return this.doc.getText(Range4.create(lineNumber - 1, index - 1, lineNumber - 1, index)).charCodeAt(0);
  }
  getText(range) {
    return this.doc.getText(range);
  }
  getPosition(offest) {
    return this.doc.positionAt(offest);
  }
};

// node_modules/yaml-language-server/lib/esm/languageservice/services/yamlValidation.js
"use strict";
var __awaiter2 = function(thisArg, _arguments, P, generator) {
  function adopt(value) {
    return value instanceof P ? value : new P(function(resolve2) {
      resolve2(value);
    });
  }
  return new (P || (P = Promise))(function(resolve2, reject) {
    function fulfilled(value) {
      try {
        step(generator.next(value));
      } catch (e) {
        reject(e);
      }
    }
    function rejected(value) {
      try {
        step(generator["throw"](value));
      } catch (e) {
        reject(e);
      }
    }
    function step(result) {
      result.done ? resolve2(result.value) : adopt(result.value).then(fulfilled, rejected);
    }
    step((generator = generator.apply(thisArg, _arguments || [])).next());
  });
};
var yamlDiagToLSDiag = (yamlDiag, textDocument) => {
  const start = textDocument.positionAt(yamlDiag.location.start);
  const range = {
    start,
    end: yamlDiag.location.toLineEnd ? Position2.create(start.line, new TextBuffer(textDocument).getLineLength(start.line)) : textDocument.positionAt(yamlDiag.location.end)
  };
  return Diagnostic3.create(range, yamlDiag.message, yamlDiag.severity, yamlDiag.code, YAML_SOURCE);
};
var YAMLValidation = class {
  constructor(schemaService) {
    this.MATCHES_MULTIPLE = "Matches multiple schemas when only one must validate.";
    this.validationEnabled = true;
    this.jsonValidation = new JSONValidation(schemaService, Promise);
  }
  configure(settings) {
    if (settings) {
      this.validationEnabled = settings.validate;
      this.customTags = settings.customTags;
      this.disableAdditionalProperties = settings.disableAdditionalProperties;
      this.yamlVersion = settings.yamlVersion;
    }
  }
  doValidation(textDocument, isKubernetes = false) {
    return __awaiter2(this, void 0, void 0, function* () {
      if (!this.validationEnabled) {
        return Promise.resolve([]);
      }
      const validationResult = [];
      try {
        const yamlDocument = yamlDocumentsCache.getYamlDocument(textDocument, { customTags: this.customTags, yamlVersion: this.yamlVersion }, true);
        let index = 0;
        for (const currentYAMLDoc of yamlDocument.documents) {
          currentYAMLDoc.isKubernetes = isKubernetes;
          currentYAMLDoc.currentDocIndex = index;
          currentYAMLDoc.disableAdditionalProperties = this.disableAdditionalProperties;
          const validation = yield this.jsonValidation.doValidation(textDocument, currentYAMLDoc);
          const syd = currentYAMLDoc;
          if (syd.errors.length > 0) {
            validationResult.push(...syd.errors);
          }
          if (syd.warnings.length > 0) {
            validationResult.push(...syd.warnings);
          }
          validationResult.push(...validation);
          index++;
        }
      } catch (err) {
        console.error(err.toString());
      }
      let previousErr;
      const foundSignatures = new Set();
      const duplicateMessagesRemoved = [];
      for (let err of validationResult) {
        if (isKubernetes && err.message === this.MATCHES_MULTIPLE) {
          continue;
        }
        if (Object.prototype.hasOwnProperty.call(err, "location")) {
          err = yamlDiagToLSDiag(err, textDocument);
        }
        if (!err.source) {
          err.source = YAML_SOURCE;
        }
        if (previousErr && previousErr.message === err.message && previousErr.range.end.line === err.range.start.line && Math.abs(previousErr.range.end.character - err.range.end.character) >= 1) {
          previousErr.range.end = err.range.end;
          continue;
        } else {
          previousErr = err;
        }
        const errSig = err.range.start.line + " " + err.range.start.character + " " + err.message;
        if (!foundSignatures.has(errSig)) {
          duplicateMessagesRemoved.push(err);
          foundSignatures.add(errSig);
        }
      }
      return duplicateMessagesRemoved;
    });
  }
};

// node_modules/yaml-language-server/lib/esm/languageservice/services/yamlFormatter.js
import { Range as Range5, Position as Position3, TextEdit as TextEdit2 } from "vscode-languageserver-types";
import {
  format as format2
} from "prettier/standalone.js";
import * as parser from "prettier/parser-yaml.js";
"use strict";
var YAMLFormatter = class {
  constructor() {
    this.formatterEnabled = true;
  }
  configure(shouldFormat) {
    if (shouldFormat) {
      this.formatterEnabled = shouldFormat.format;
    }
  }
  format(document, options) {
    if (!this.formatterEnabled) {
      return [];
    }
    try {
      const text = document.getText();
      const prettierOptions = {
        parser: "yaml",
        plugins: [parser],
        tabWidth: options.tabWidth || options.tabSize,
        singleQuote: options.singleQuote,
        bracketSpacing: options.bracketSpacing,
        proseWrap: options.proseWrap === "always" ? "always" : options.proseWrap === "never" ? "never" : "preserve",
        printWidth: options.printWidth
      };
      const formatted = format2(text, prettierOptions);
      return [TextEdit2.replace(Range5.create(Position3.create(0, 0), document.positionAt(text.length)), formatted)];
    } catch (error) {
      return [];
    }
  }
};

// node_modules/yaml-language-server/lib/esm/languageservice/services/yamlLinks.js
function findLinks2(document) {
  try {
    const doc = yamlDocumentsCache.getYamlDocument(document);
    const linkPromises = [];
    for (const yamlDoc of doc.documents) {
      linkPromises.push(findLinks(document, yamlDoc));
    }
    return Promise.all(linkPromises).then((yamlLinkArray) => [].concat(...yamlLinkArray));
  } catch (err) {
    this.telemetry.sendError("yaml.documentLink.error", { error: err });
  }
}

// node_modules/yaml-language-server/lib/esm/languageservice/services/yamlFolding.js
import { FoldingRange as FoldingRange2, Range as Range6 } from "vscode-languageserver-types";
function getFoldingRanges2(document, context) {
  if (!document) {
    return;
  }
  const result = [];
  const doc = yamlDocumentsCache.getYamlDocument(document);
  for (const ymlDoc of doc.documents) {
    ymlDoc.visit((node) => {
      var _a;
      if (node.type === "property" && node.valueNode.type === "array" || node.type === "object" && ((_a = node.parent) === null || _a === void 0 ? void 0 : _a.type) === "array") {
        result.push(creteNormalizedFolding(document, node));
      }
      if (node.type === "property" && node.valueNode.type === "object") {
        result.push(creteNormalizedFolding(document, node));
      }
      return true;
    });
  }
  const rangeLimit = context && context.rangeLimit;
  if (typeof rangeLimit !== "number" || result.length <= rangeLimit) {
    return result;
  }
  if (context && context.onRangeLimitExceeded) {
    context.onRangeLimitExceeded(document.uri);
  }
  return result.slice(0, context.rangeLimit);
}
function creteNormalizedFolding(document, node) {
  const startPos = document.positionAt(node.offset);
  let endPos = document.positionAt(node.offset + node.length);
  const textFragment = document.getText(Range6.create(startPos, endPos));
  const newLength = textFragment.length - textFragment.trimRight().length;
  if (newLength > 0) {
    endPos = document.positionAt(node.offset + node.length - newLength);
  }
  return FoldingRange2.create(startPos.line, endPos.line, startPos.character, endPos.character);
}

// node_modules/yaml-language-server/lib/esm/languageservice/services/yamlCodeActions.js
import { CodeAction as CodeAction2, CodeActionKind as CodeActionKind2, Command as Command2, Position as Position4, Range as Range7, TextEdit as TextEdit3 } from "vscode-languageserver-types";

// node_modules/yaml-language-server/lib/esm/commands.js
var YamlCommands;
(function(YamlCommands2) {
  YamlCommands2["JUMP_TO_SCHEMA"] = "jumpToSchema";
})(YamlCommands || (YamlCommands = {}));

// node_modules/yaml-language-server/lib/esm/languageservice/services/yamlCodeActions.js
import {
  basename as basename2
} from "path-browserify";
var YamlCodeActions = class {
  constructor(clientCapabilities) {
    this.clientCapabilities = clientCapabilities;
    this.indentation = "  ";
  }
  configure(settings) {
    this.indentation = settings.indentation;
  }
  getCodeAction(document, params) {
    if (!params.context.diagnostics) {
      return;
    }
    const result = [];
    result.push(...this.getJumpToSchemaActions(params.context.diagnostics));
    result.push(...this.getTabToSpaceConverting(params.context.diagnostics, document));
    return result;
  }
  getJumpToSchemaActions(diagnostics) {
    var _a, _b, _c, _d, _e;
    const isOpenTextDocumentEnabled = (_d = (_c = (_b = (_a = this.clientCapabilities) === null || _a === void 0 ? void 0 : _a.window) === null || _b === void 0 ? void 0 : _b.showDocument) === null || _c === void 0 ? void 0 : _c.support) !== null && _d !== void 0 ? _d : false;
    if (!isOpenTextDocumentEnabled) {
      return [];
    }
    const schemaUriToDiagnostic = new Map();
    for (const diagnostic of diagnostics) {
      const schemaUri = ((_e = diagnostic.data) === null || _e === void 0 ? void 0 : _e.schemaUri) || [];
      for (const schemaUriStr of schemaUri) {
        if (schemaUriStr) {
          if (!schemaUriToDiagnostic.has(schemaUriStr)) {
            schemaUriToDiagnostic.set(schemaUriStr, []);
          }
          schemaUriToDiagnostic.get(schemaUriStr).push(diagnostic);
        }
      }
    }
    const result = [];
    for (const schemaUri of schemaUriToDiagnostic.keys()) {
      const action = CodeAction2.create(`Jump to schema location (${basename2(schemaUri)})`, Command2.create("JumpToSchema", YamlCommands.JUMP_TO_SCHEMA, schemaUri));
      action.diagnostics = schemaUriToDiagnostic.get(schemaUri);
      result.push(action);
    }
    return result;
  }
  getTabToSpaceConverting(diagnostics, document) {
    const result = [];
    const textBuff = new TextBuffer(document);
    const processedLine = [];
    for (const diag of diagnostics) {
      if (diag.message === "Using tabs can lead to unpredictable results") {
        if (processedLine.includes(diag.range.start.line)) {
          continue;
        }
        const lineContent = textBuff.getLineContent(diag.range.start.line);
        let replacedTabs = 0;
        let newText = "";
        for (let i = diag.range.start.character; i <= diag.range.end.character; i++) {
          const char = lineContent.charAt(i);
          if (char !== "	") {
            break;
          }
          replacedTabs++;
          newText += this.indentation;
        }
        processedLine.push(diag.range.start.line);
        let resultRange = diag.range;
        if (replacedTabs !== diag.range.end.character - diag.range.start.character) {
          resultRange = Range7.create(diag.range.start, Position4.create(diag.range.end.line, diag.range.start.character + replacedTabs));
        }
        result.push(CodeAction2.create("Convert Tab to Spaces", createWorkspaceEdit(document.uri, [TextEdit3.replace(resultRange, newText)]), CodeActionKind2.QuickFix));
      }
    }
    if (result.length !== 0) {
      const replaceEdits = [];
      for (let i = 0; i <= textBuff.getLineCount(); i++) {
        const lineContent = textBuff.getLineContent(i);
        let replacedTabs = 0;
        let newText = "";
        for (let j = 0; j < lineContent.length; j++) {
          const char = lineContent.charAt(j);
          if (char !== " " && char !== "	") {
            if (replacedTabs !== 0) {
              replaceEdits.push(TextEdit3.replace(Range7.create(i, j - replacedTabs, i, j), newText));
              replacedTabs = 0;
              newText = "";
            }
            break;
          }
          if (char === " " && replacedTabs !== 0) {
            replaceEdits.push(TextEdit3.replace(Range7.create(i, j - replacedTabs, i, j), newText));
            replacedTabs = 0;
            newText = "";
            continue;
          }
          if (char === "	") {
            newText += this.indentation;
            replacedTabs++;
          }
        }
        if (replacedTabs !== 0) {
          replaceEdits.push(TextEdit3.replace(Range7.create(i, 0, i, textBuff.getLineLength(i)), newText));
        }
      }
      if (replaceEdits.length > 0) {
        result.push(CodeAction2.create("Convert all Tabs to Spaces", createWorkspaceEdit(document.uri, replaceEdits), CodeActionKind2.QuickFix));
      }
    }
    return result;
  }
};
function createWorkspaceEdit(uri, edits) {
  const changes = {};
  changes[uri] = edits;
  const edit = {
    changes
  };
  return edit;
}

// node_modules/yaml-language-server/lib/esm/languageserver/commandExecutor.js
var CommandExecutor = class {
  constructor() {
    this.commands = new Map();
  }
  executeCommand(params) {
    if (this.commands.has(params.command)) {
      const handler = this.commands.get(params.command);
      return handler(...params.arguments);
    }
    throw new Error(`Command '${params.command}' not found`);
  }
  registerCommand(commandId, handler) {
    this.commands.set(commandId, handler);
  }
};
var commandExecutor = new CommandExecutor();

// node_modules/yaml-language-server/lib/esm/languageservice/services/yamlOnTypeFormatting.js
import { Position as Position5, Range as Range8, TextEdit as TextEdit4 } from "vscode-languageserver-types";
function doDocumentOnTypeFormatting(document, params) {
  const { position } = params;
  const tb = new TextBuffer(document);
  if (params.ch === "\n") {
    const previousLine = tb.getLineContent(position.line - 1);
    if (previousLine.trimRight().endsWith(":")) {
      const currentLine = tb.getLineContent(position.line);
      const subLine = currentLine.substring(position.character, currentLine.length);
      const isInArray = previousLine.indexOf(" - ") !== -1;
      if (subLine.trimRight().length === 0) {
        const indentationFix = position.character - (previousLine.length - previousLine.trimLeft().length);
        if (indentationFix === params.options.tabSize && !isInArray) {
          return;
        }
        const result = [];
        if (currentLine.length > 0) {
          result.push(TextEdit4.del(Range8.create(position, Position5.create(position.line, currentLine.length - 1))));
        }
        result.push(TextEdit4.insert(position, " ".repeat(params.options.tabSize + (isInArray ? 2 - indentationFix : 0))));
        return result;
      }
      if (isInArray) {
        return [TextEdit4.insert(position, " ".repeat(params.options.tabSize))];
      }
    }
    if (previousLine.trimRight().endsWith("|")) {
      return [TextEdit4.insert(position, " ".repeat(params.options.tabSize))];
    }
    if (previousLine.includes(" - ") && !previousLine.includes(": ")) {
      return [TextEdit4.insert(position, "- ")];
    }
    if (previousLine.includes(" - ") && previousLine.includes(": ")) {
      return [TextEdit4.insert(position, "  ")];
    }
  }
  return;
}

// node_modules/yaml-language-server/lib/esm/languageservice/services/yamlCodeLens.js
import { CodeLens, Range as Range9 } from "vscode-languageserver-types";
import {
  basename as basename3,
  extname
} from "path-browserify";
var __awaiter3 = function(thisArg, _arguments, P, generator) {
  function adopt(value) {
    return value instanceof P ? value : new P(function(resolve2) {
      resolve2(value);
    });
  }
  return new (P || (P = Promise))(function(resolve2, reject) {
    function fulfilled(value) {
      try {
        step(generator.next(value));
      } catch (e) {
        reject(e);
      }
    }
    function rejected(value) {
      try {
        step(generator["throw"](value));
      } catch (e) {
        reject(e);
      }
    }
    function step(result) {
      result.done ? resolve2(result.value) : adopt(result.value).then(fulfilled, rejected);
    }
    step((generator = generator.apply(thisArg, _arguments || [])).next());
  });
};
var YamlCodeLens = class {
  constructor(schemaService, telemetry) {
    this.schemaService = schemaService;
    this.telemetry = telemetry;
  }
  getCodeLens(document, params) {
    return __awaiter3(this, void 0, void 0, function* () {
      const result = [];
      try {
        const yamlDocument = yamlDocumentsCache.getYamlDocument(document);
        for (const currentYAMLDoc of yamlDocument.documents) {
          const schema = yield this.schemaService.getSchemaForResource(document.uri, currentYAMLDoc);
          if (schema === null || schema === void 0 ? void 0 : schema.schema) {
            const schemaUrls = getSchemaUrl(schema === null || schema === void 0 ? void 0 : schema.schema);
            if (schemaUrls.size === 0) {
              continue;
            }
            for (const urlToSchema of schemaUrls) {
              const lens = CodeLens.create(Range9.create(0, 0, 0, 0));
              lens.command = {
                title: getCommandTitle(urlToSchema[0], urlToSchema[1]),
                command: YamlCommands.JUMP_TO_SCHEMA,
                arguments: [urlToSchema[0]]
              };
              result.push(lens);
            }
          }
        }
      } catch (err) {
        this.telemetry.sendError("yaml.codeLens.error", { error: err, documentUri: document.uri });
      }
      return result;
    });
  }
  resolveCodeLens(param) {
    return param;
  }
};
function getCommandTitle(url, schema) {
  const uri = URI.parse(url);
  let baseName = basename3(uri.fsPath);
  if (!extname(uri.fsPath)) {
    baseName += ".json";
  }
  if (Object.getOwnPropertyDescriptor(schema, "name")) {
    return Object.getOwnPropertyDescriptor(schema, "name").value + ` (${baseName})`;
  } else if (schema.title) {
    return schema.title + ` (${baseName})`;
  }
  return baseName;
}
function getSchemaUrl(schema) {
  const result = new Map();
  if (!schema) {
    return result;
  }
  const url = schema.url;
  if (url) {
    if (url.startsWith("schemaservice://combinedSchema/")) {
      addSchemasForOf(schema, result);
    } else {
      result.set(schema.url, schema);
    }
  } else {
    addSchemasForOf(schema, result);
  }
  return result;
}
function addSchemasForOf(schema, result) {
  if (schema.allOf) {
    addInnerSchemaUrls(schema.allOf, result);
  }
  if (schema.anyOf) {
    addInnerSchemaUrls(schema.anyOf, result);
  }
  if (schema.oneOf) {
    addInnerSchemaUrls(schema.oneOf, result);
  }
}
function addInnerSchemaUrls(schemas, result) {
  for (const subSchema of schemas) {
    if (!isBoolean2(subSchema)) {
      if (subSchema.url && !result.has(subSchema.url)) {
        result.set(subSchema.url, subSchema);
      }
    }
  }
}

// node_modules/yaml-language-server/lib/esm/languageservice/services/yamlCommands.js
var __awaiter4 = function(thisArg, _arguments, P, generator) {
  function adopt(value) {
    return value instanceof P ? value : new P(function(resolve2) {
      resolve2(value);
    });
  }
  return new (P || (P = Promise))(function(resolve2, reject) {
    function fulfilled(value) {
      try {
        step(generator.next(value));
      } catch (e) {
        reject(e);
      }
    }
    function rejected(value) {
      try {
        step(generator["throw"](value));
      } catch (e) {
        reject(e);
      }
    }
    function step(result) {
      result.done ? resolve2(result.value) : adopt(result.value).then(fulfilled, rejected);
    }
    step((generator = generator.apply(thisArg, _arguments || [])).next());
  });
};
function registerCommands(commandExecutor2, connection) {
  commandExecutor2.registerCommand(YamlCommands.JUMP_TO_SCHEMA, (uri) => __awaiter4(this, void 0, void 0, function* () {
    if (!uri) {
      return;
    }
    if (!uri.startsWith("file") && !/^[a-z]:[\\/]/i.test(uri)) {
      const origUri = URI.parse(uri);
      const customUri = URI.from({
        scheme: "json-schema",
        authority: origUri.authority,
        path: origUri.path.endsWith(".json") ? origUri.path : origUri.path + ".json",
        fragment: uri
      });
      uri = customUri.toString();
    }
    if (/^[a-z]:[\\/]/i.test(uri)) {
      const winUri = URI.file(uri);
      uri = winUri.toString();
    }
    const result = yield connection.window.showDocument({ uri, external: false, takeFocus: true });
    if (!result) {
      connection.window.showErrorMessage(`Cannot open ${uri}`);
    }
  }));
}

// node_modules/yaml-language-server/lib/esm/languageservice/services/yamlCompletion.js
import { CompletionItem as CompletionItem2, CompletionItemKind as CompletionItemKind2, CompletionList as CompletionList2, InsertTextFormat as InsertTextFormat2, InsertTextMode, MarkupKind as MarkupKind2, Range as Range10, TextEdit as TextEdit5 } from "vscode-languageserver-types";
import { isPair as isPair3, isScalar as isScalar4, isMap as isMap3, isSeq as isSeq3, isNode as isNode2 } from "yaml";

// node_modules/yaml-language-server/lib/esm/languageservice/utils/indentationGuesser.js
var SpacesDiffResult = class {
  constructor() {
    this.spacesDiff = 0;
    this.looksLikeAlignment = false;
  }
};
function spacesDiff(a2, aLength, b, bLength, result) {
  result.spacesDiff = 0;
  result.looksLikeAlignment = false;
  let i;
  for (i = 0; i < aLength && i < bLength; i++) {
    const aCharCode = a2.charCodeAt(i);
    const bCharCode = b.charCodeAt(i);
    if (aCharCode !== bCharCode) {
      break;
    }
  }
  let aSpacesCnt = 0, aTabsCount = 0;
  for (let j = i; j < aLength; j++) {
    const aCharCode = a2.charCodeAt(j);
    if (aCharCode === 32) {
      aSpacesCnt++;
    } else {
      aTabsCount++;
    }
  }
  let bSpacesCnt = 0, bTabsCount = 0;
  for (let j = i; j < bLength; j++) {
    const bCharCode = b.charCodeAt(j);
    if (bCharCode === 32) {
      bSpacesCnt++;
    } else {
      bTabsCount++;
    }
  }
  if (aSpacesCnt > 0 && aTabsCount > 0) {
    return;
  }
  if (bSpacesCnt > 0 && bTabsCount > 0) {
    return;
  }
  const tabsDiff = Math.abs(aTabsCount - bTabsCount);
  const spacesDiff2 = Math.abs(aSpacesCnt - bSpacesCnt);
  if (tabsDiff === 0) {
    result.spacesDiff = spacesDiff2;
    if (spacesDiff2 > 0 && 0 <= bSpacesCnt - 1 && bSpacesCnt - 1 < a2.length && bSpacesCnt < b.length) {
      if (b.charCodeAt(bSpacesCnt) !== 32 && a2.charCodeAt(bSpacesCnt - 1) === 32) {
        if (a2.charCodeAt(a2.length - 1) === 44) {
          result.looksLikeAlignment = true;
        }
      }
    }
    return;
  }
  if (spacesDiff2 % tabsDiff === 0) {
    result.spacesDiff = spacesDiff2 / tabsDiff;
    return;
  }
}
function guessIndentation(source, defaultTabSize, defaultInsertSpaces) {
  const linesCount = Math.min(source.getLineCount(), 1e4);
  let linesIndentedWithTabsCount = 0;
  let linesIndentedWithSpacesCount = 0;
  let previousLineText = "";
  let previousLineIndentation = 0;
  const ALLOWED_TAB_SIZE_GUESSES = [2, 4, 6, 8, 3, 5, 7];
  const MAX_ALLOWED_TAB_SIZE_GUESS = 8;
  const spacesDiffCount = [0, 0, 0, 0, 0, 0, 0, 0, 0];
  const tmp = new SpacesDiffResult();
  for (let lineNumber = 1; lineNumber <= linesCount; lineNumber++) {
    const currentLineLength = source.getLineLength(lineNumber);
    const currentLineText = source.getLineContent(lineNumber);
    const useCurrentLineText = currentLineLength <= 65536;
    let currentLineHasContent = false;
    let currentLineIndentation = 0;
    let currentLineSpacesCount = 0;
    let currentLineTabsCount = 0;
    for (let j = 0, lenJ = currentLineLength; j < lenJ; j++) {
      const charCode = useCurrentLineText ? currentLineText.charCodeAt(j) : source.getLineCharCode(lineNumber, j);
      if (charCode === 9) {
        currentLineTabsCount++;
      } else if (charCode === 32) {
        currentLineSpacesCount++;
      } else {
        currentLineHasContent = true;
        currentLineIndentation = j;
        break;
      }
    }
    if (!currentLineHasContent) {
      continue;
    }
    if (currentLineTabsCount > 0) {
      linesIndentedWithTabsCount++;
    } else if (currentLineSpacesCount > 1) {
      linesIndentedWithSpacesCount++;
    }
    spacesDiff(previousLineText, previousLineIndentation, currentLineText, currentLineIndentation, tmp);
    if (tmp.looksLikeAlignment) {
      if (!(defaultInsertSpaces && defaultTabSize === tmp.spacesDiff)) {
        continue;
      }
    }
    const currentSpacesDiff = tmp.spacesDiff;
    if (currentSpacesDiff <= MAX_ALLOWED_TAB_SIZE_GUESS) {
      spacesDiffCount[currentSpacesDiff]++;
    }
    previousLineText = currentLineText;
    previousLineIndentation = currentLineIndentation;
  }
  let insertSpaces = defaultInsertSpaces;
  if (linesIndentedWithTabsCount !== linesIndentedWithSpacesCount) {
    insertSpaces = linesIndentedWithTabsCount < linesIndentedWithSpacesCount;
  }
  let tabSize = defaultTabSize;
  if (insertSpaces) {
    let tabSizeScore = insertSpaces ? 0 : 0.1 * linesCount;
    ALLOWED_TAB_SIZE_GUESSES.forEach((possibleTabSize) => {
      const possibleTabSizeScore = spacesDiffCount[possibleTabSize];
      if (possibleTabSizeScore > tabSizeScore) {
        tabSizeScore = possibleTabSizeScore;
        tabSize = possibleTabSize;
      }
    });
    if (tabSize === 4 && spacesDiffCount[4] > 0 && spacesDiffCount[2] > 0 && spacesDiffCount[2] >= spacesDiffCount[4] / 2) {
      tabSize = 2;
    }
  }
  return {
    insertSpaces,
    tabSize
  };
}

// node_modules/yaml-language-server/lib/esm/languageservice/utils/json.js
function stringifyObject2(obj, indent, stringifyLiteral, settings, depth = 0, consecutiveArrays = 0) {
  if (obj !== null && typeof obj === "object") {
    const newIndent = depth === 0 && settings.shouldIndentWithTab || depth > 0 ? indent + "  " : "";
    if (Array.isArray(obj)) {
      consecutiveArrays += 1;
      if (obj.length === 0) {
        return "";
      }
      let result = "";
      for (let i = 0; i < obj.length; i++) {
        let pseudoObj = obj[i];
        if (!Array.isArray(obj[i])) {
          pseudoObj = preprendToObject(obj[i], consecutiveArrays);
        }
        result += newIndent + stringifyObject2(pseudoObj, indent, stringifyLiteral, settings, depth += 1, consecutiveArrays);
        if (i < obj.length - 1) {
          result += "\n";
        }
      }
      result += indent;
      return result;
    } else {
      const keys = Object.keys(obj);
      if (keys.length === 0) {
        return "";
      }
      let result = depth === 0 && settings.newLineFirst || depth > 0 ? "\n" : "";
      for (let i = 0; i < keys.length; i++) {
        const key = keys[i];
        if (depth === 0 && i === 0 && !settings.indentFirstObject) {
          result += indent + key + ": " + stringifyObject2(obj[key], newIndent, stringifyLiteral, settings, depth += 1, 0);
        } else {
          result += newIndent + key + ": " + stringifyObject2(obj[key], newIndent, stringifyLiteral, settings, depth += 1, 0);
        }
        if (i < keys.length - 1) {
          result += "\n";
        }
      }
      result += indent;
      return result;
    }
  }
  return stringifyLiteral(obj);
}
function preprendToObject(obj, consecutiveArrays) {
  const newObj = {};
  for (let i = 0; i < Object.keys(obj).length; i++) {
    const key = Object.keys(obj)[i];
    if (i === 0) {
      newObj["- ".repeat(consecutiveArrays) + key] = obj[key];
    } else {
      newObj["  ".repeat(consecutiveArrays) + key] = obj[key];
    }
  }
  return newObj;
}

// node_modules/yaml-language-server/lib/esm/languageservice/services/yamlCompletion.js
var __awaiter5 = function(thisArg, _arguments, P, generator) {
  function adopt(value) {
    return value instanceof P ? value : new P(function(resolve2) {
      resolve2(value);
    });
  }
  return new (P || (P = Promise))(function(resolve2, reject) {
    function fulfilled(value) {
      try {
        step(generator.next(value));
      } catch (e) {
        reject(e);
      }
    }
    function rejected(value) {
      try {
        step(generator["throw"](value));
      } catch (e) {
        reject(e);
      }
    }
    function step(result) {
      result.done ? resolve2(result.value) : adopt(result.value).then(fulfilled, rejected);
    }
    step((generator = generator.apply(thisArg, _arguments || [])).next());
  });
};
var localize9 = loadMessageBundle();
var doubleQuotesEscapeRegExp = /[\\]+"/g;
var YamlCompletion = class {
  constructor(schemaService, clientCapabilities = {}, yamlDocument, telemetry) {
    this.schemaService = schemaService;
    this.clientCapabilities = clientCapabilities;
    this.yamlDocument = yamlDocument;
    this.telemetry = telemetry;
    this.completionEnabled = true;
  }
  configure(languageSettings) {
    if (languageSettings) {
      this.completionEnabled = languageSettings.completion;
    }
    this.customTags = languageSettings.customTags;
    this.yamlVersion = languageSettings.yamlVersion;
    this.configuredIndentation = languageSettings.indentation;
  }
  doComplete(document, position, isKubernetes = false) {
    return __awaiter5(this, void 0, void 0, function* () {
      const result = CompletionList2.create([], false);
      if (!this.completionEnabled) {
        return result;
      }
      const doc = this.yamlDocument.getYamlDocument(document, { customTags: this.customTags, yamlVersion: this.yamlVersion }, true);
      const textBuffer = new TextBuffer(document);
      if (!this.configuredIndentation) {
        const indent = guessIndentation(textBuffer, 2, true);
        this.indentation = indent.insertSpaces ? " ".repeat(indent.tabSize) : "	";
      } else {
        this.indentation = this.configuredIndentation;
      }
      setKubernetesParserOption(doc.documents, isKubernetes);
      const offset = document.offsetAt(position);
      if (document.getText().charAt(offset - 1) === ":") {
        return Promise.resolve(result);
      }
      const currentDoc = matchOffsetToDocument(offset, doc);
      if (currentDoc === null) {
        return Promise.resolve(result);
      }
      let [node, foundByClosest] = currentDoc.getNodeFromPosition(offset, textBuffer);
      const currentWord = this.getCurrentWord(document, offset);
      let overwriteRange = null;
      if (node && isScalar4(node) && node.value === "null") {
        const nodeStartPos = document.positionAt(node.range[0]);
        nodeStartPos.character += 1;
        const nodeEndPos = document.positionAt(node.range[2]);
        nodeEndPos.character += 1;
        overwriteRange = Range10.create(nodeStartPos, nodeEndPos);
      } else if (node && isScalar4(node) && node.value) {
        const start = document.positionAt(node.range[0]);
        if (offset > 0 && start.character > 0 && document.getText().charAt(offset - 1) === "-") {
          start.character -= 1;
        }
        overwriteRange = Range10.create(start, document.positionAt(node.range[1]));
      } else {
        let overwriteStart = document.offsetAt(position) - currentWord.length;
        if (overwriteStart > 0 && document.getText()[overwriteStart - 1] === '"') {
          overwriteStart--;
        }
        overwriteRange = Range10.create(document.positionAt(overwriteStart), position);
      }
      const proposed = {};
      const collector = {
        add: (completionItem) => {
          let label = completionItem.label;
          if (!label) {
            console.warn(`Ignoring CompletionItem without label: ${JSON.stringify(completionItem)}`);
            return;
          }
          if (!isString2(label)) {
            label = String(label);
          }
          const existing = proposed[label];
          if (!existing) {
            label = label.replace(/[\n]/g, "\u21B5");
            if (label.length > 60) {
              const shortendedLabel = label.substr(0, 57).trim() + "...";
              if (!proposed[shortendedLabel]) {
                label = shortendedLabel;
              }
            }
            if (overwriteRange && overwriteRange.start.line === overwriteRange.end.line) {
              completionItem.textEdit = TextEdit5.replace(overwriteRange, completionItem.insertText);
            }
            completionItem.label = label;
            proposed[label] = completionItem;
            result.items.push(completionItem);
          }
        },
        error: (message) => {
          console.error(message);
          this.telemetry.sendError("yaml.completion.error", { error: message });
        },
        log: (message) => {
          console.log(message);
        },
        getNumberOfProposals: () => {
          return result.items.length;
        }
      };
      if (this.customTags.length > 0) {
        this.getCustomTagValueCompletions(collector);
      }
      let lineContent = textBuffer.getLineContent(position.line);
      if (lineContent.endsWith("\n")) {
        lineContent = lineContent.substr(0, lineContent.length - 1);
      }
      try {
        const schema = yield this.schemaService.getSchemaForResource(document.uri, currentDoc);
        if (!schema || schema.errors.length) {
          if (position.line === 0 && position.character === 0 && !isModeline(lineContent)) {
            const inlineSchemaCompletion = {
              kind: CompletionItemKind2.Text,
              label: "Inline schema",
              insertText: "# yaml-language-server: $schema=",
              insertTextFormat: InsertTextFormat2.PlainText
            };
            result.items.push(inlineSchemaCompletion);
          }
        }
        if (isModeline(lineContent) || isInComment(doc.tokens, offset)) {
          const schemaIndex = lineContent.indexOf("$schema=");
          if (schemaIndex !== -1 && schemaIndex + "$schema=".length <= position.character) {
            this.schemaService.getAllSchemas().forEach((schema2) => {
              var _a;
              const schemaIdCompletion = {
                kind: CompletionItemKind2.Constant,
                label: (_a = schema2.name) !== null && _a !== void 0 ? _a : schema2.uri,
                detail: schema2.description,
                insertText: schema2.uri,
                insertTextFormat: InsertTextFormat2.PlainText,
                insertTextMode: InsertTextMode.asIs
              };
              result.items.push(schemaIdCompletion);
            });
          }
          return result;
        }
        if (!schema || schema.errors.length) {
          return result;
        }
        let currentProperty = null;
        if (!node) {
          if (!currentDoc.internalDocument.contents || isScalar4(currentDoc.internalDocument.contents)) {
            const map = currentDoc.internalDocument.createNode({});
            map.range = [offset, offset + 1, offset + 1];
            currentDoc.internalDocument.contents = map;
            currentDoc.internalDocument = currentDoc.internalDocument;
            node = map;
          } else {
            node = currentDoc.findClosestNode(offset, textBuffer);
            foundByClosest = true;
          }
        }
        if (node) {
          if (lineContent.length === 0) {
            node = currentDoc.internalDocument.contents;
          } else {
            const parent = currentDoc.getParent(node);
            if (parent) {
              if (isScalar4(node)) {
                if (node.value) {
                  if (isPair3(parent)) {
                    if (parent.value === node) {
                      if (lineContent.trim().length > 0 && lineContent.indexOf(":") < 0) {
                        const map = this.createTempObjNode(currentWord, node, currentDoc);
                        if (isSeq3(currentDoc.internalDocument.contents)) {
                          const index = indexOf(currentDoc.internalDocument.contents, parent);
                          if (typeof index === "number") {
                            currentDoc.internalDocument.set(index, map);
                            currentDoc.internalDocument = currentDoc.internalDocument;
                          }
                        } else {
                          currentDoc.internalDocument.set(parent.key, map);
                          currentDoc.internalDocument = currentDoc.internalDocument;
                        }
                        currentProperty = map.items[0];
                        node = map;
                      } else if (lineContent.trim().length === 0) {
                        const parentParent = currentDoc.getParent(parent);
                        if (parentParent) {
                          node = parentParent;
                        }
                      }
                    } else if (parent.key === node) {
                      const parentParent = currentDoc.getParent(parent);
                      currentProperty = parent;
                      if (parentParent) {
                        node = parentParent;
                      }
                    }
                  } else if (isSeq3(parent)) {
                    if (lineContent.trim().length > 0) {
                      const map = this.createTempObjNode(currentWord, node, currentDoc);
                      parent.delete(node);
                      parent.add(map);
                      currentDoc.internalDocument = currentDoc.internalDocument;
                      node = map;
                    } else {
                      node = parent;
                    }
                  }
                } else if (node.value === null) {
                  if (isPair3(parent)) {
                    if (parent.key === node) {
                      node = parent;
                    } else {
                      if (isNode2(parent.key) && parent.key.range) {
                        const parentParent = currentDoc.getParent(parent);
                        if (foundByClosest && parentParent && isMap3(parentParent) && isMapContainsEmptyPair(parentParent)) {
                          node = parentParent;
                        } else {
                          const parentPosition = document.positionAt(parent.key.range[0]);
                          if (position.character > parentPosition.character && position.line !== parentPosition.line) {
                            const map = this.createTempObjNode(currentWord, node, currentDoc);
                            if (parentParent && (isMap3(parentParent) || isSeq3(parentParent))) {
                              parentParent.set(parent.key, map);
                              currentDoc.internalDocument = currentDoc.internalDocument;
                            } else {
                              currentDoc.internalDocument.set(parent.key, map);
                              currentDoc.internalDocument = currentDoc.internalDocument;
                            }
                            currentProperty = map.items[0];
                            node = map;
                          } else if (parentPosition.character === position.character) {
                            if (parentParent) {
                              node = parentParent;
                            }
                          }
                        }
                      }
                    }
                  } else if (isSeq3(parent)) {
                    if (lineContent.charAt(position.character - 1) !== "-") {
                      const map = this.createTempObjNode(currentWord, node, currentDoc);
                      parent.delete(node);
                      parent.add(map);
                      currentDoc.internalDocument = currentDoc.internalDocument;
                      node = map;
                    } else {
                      node = parent;
                    }
                  }
                }
              } else if (isMap3(node)) {
                if (!foundByClosest && lineContent.trim().length === 0 && isSeq3(parent)) {
                  const nextLine = textBuffer.getLineContent(position.line + 1);
                  if (textBuffer.getLineCount() === position.line + 1 || nextLine.trim().length === 0) {
                    node = parent;
                  }
                }
              }
            } else if (isScalar4(node)) {
              const map = this.createTempObjNode(currentWord, node, currentDoc);
              currentDoc.internalDocument.contents = map;
              currentDoc.internalDocument = currentDoc.internalDocument;
              currentProperty = map.items[0];
              node = map;
            } else if (isMap3(node)) {
              for (const pair of node.items) {
                if (isNode2(pair.value) && pair.value.range && pair.value.range[0] === offset + 1) {
                  node = pair.value;
                }
              }
            }
          }
        }
        if (node && isMap3(node)) {
          const properties = node.items;
          for (const p of properties) {
            if (!currentProperty || currentProperty !== p) {
              if (isScalar4(p.key)) {
                proposed[p.key.value.toString()] = CompletionItem2.create("__");
              }
            }
          }
          this.addPropertyCompletions(schema, currentDoc, node, "", collector, textBuffer, overwriteRange);
          if (!schema && currentWord.length > 0 && document.getText().charAt(offset - currentWord.length - 1) !== '"') {
            collector.add({
              kind: CompletionItemKind2.Property,
              label: currentWord,
              insertText: this.getInsertTextForProperty(currentWord, null, ""),
              insertTextFormat: InsertTextFormat2.Snippet
            });
          }
        }
        const types = {};
        this.getValueCompletions(schema, currentDoc, node, offset, document, collector, types);
      } catch (err) {
        if (err.stack) {
          console.error(err.stack);
        } else {
          console.error(err);
        }
        this.telemetry.sendError("yaml.completion.error", { error: err });
      }
      return result;
    });
  }
  createTempObjNode(currentWord, node, currentDoc) {
    const obj = {};
    obj[currentWord] = null;
    const map = currentDoc.internalDocument.createNode(obj);
    map.range = node.range;
    map.items[0].key.range = node.range;
    map.items[0].value.range = node.range;
    return map;
  }
  addPropertyCompletions(schema, doc, node, separatorAfter, collector, textBuffer, overwriteRange) {
    const matchingSchemas = doc.getMatchingSchemas(schema.schema);
    const existingKey = textBuffer.getText(overwriteRange);
    const hasColumn = textBuffer.getLineContent(overwriteRange.start.line).indexOf(":") === -1;
    const nodeParent = doc.getParent(node);
    for (const schema2 of matchingSchemas) {
      if (schema2.node.internalNode === node && !schema2.inverted) {
        this.collectDefaultSnippets(schema2.schema, separatorAfter, collector, {
          newLineFirst: false,
          indentFirstObject: false,
          shouldIndentWithTab: false
        });
        const schemaProperties = schema2.schema.properties;
        if (schemaProperties) {
          const maxProperties = schema2.schema.maxProperties;
          if (maxProperties === void 0 || node.items === void 0 || node.items.length < maxProperties || isMapContainsEmptyPair(node)) {
            for (const key in schemaProperties) {
              if (Object.prototype.hasOwnProperty.call(schemaProperties, key)) {
                const propertySchema = schemaProperties[key];
                if (typeof propertySchema === "object" && !propertySchema.deprecationMessage && !propertySchema["doNotSuggest"]) {
                  let identCompensation = "";
                  if (nodeParent && isSeq3(nodeParent) && node.items.length <= 1) {
                    const sourceText = textBuffer.getText();
                    const indexOfSlash = sourceText.lastIndexOf("-", node.range[0] - 1);
                    if (indexOfSlash >= 0) {
                      identCompensation = " " + sourceText.slice(indexOfSlash + 1, node.range[0]);
                    }
                  }
                  let pair;
                  if (propertySchema.type === "array" && (pair = node.items.find((it) => isScalar4(it.key) && it.key.range && it.key.value === key && isScalar4(it.value) && !it.value.value && textBuffer.getPosition(it.key.range[2]).line === overwriteRange.end.line - 1)) && pair) {
                    if (Array.isArray(propertySchema.items)) {
                      this.addSchemaValueCompletions(propertySchema.items[0], separatorAfter, collector, {});
                    } else if (typeof propertySchema.items === "object" && propertySchema.items.type === "object") {
                      collector.add({
                        kind: this.getSuggestionKind(propertySchema.items.type),
                        label: "- (array item)",
                        documentation: `Create an item of an array${propertySchema.description === void 0 ? "" : "(" + propertySchema.description + ")"}`,
                        insertText: `- ${this.getInsertTextForObject(propertySchema.items, separatorAfter, "  ").insertText.trimLeft()}`,
                        insertTextFormat: InsertTextFormat2.Snippet
                      });
                    }
                  }
                  let insertText = key;
                  if (!key.startsWith(existingKey) || hasColumn) {
                    insertText = this.getInsertTextForProperty(key, propertySchema, separatorAfter, identCompensation + this.indentation);
                  }
                  collector.add({
                    kind: CompletionItemKind2.Property,
                    label: key,
                    insertText,
                    insertTextFormat: InsertTextFormat2.Snippet,
                    documentation: this.fromMarkup(propertySchema.markdownDescription) || propertySchema.description || ""
                  });
                }
              }
            }
          }
        }
        if (nodeParent && isSeq3(nodeParent) && schema2.schema.type !== "object") {
          this.addSchemaValueCompletions(schema2.schema, separatorAfter, collector, {});
        }
      }
      if (nodeParent && schema2.node.internalNode === nodeParent && schema2.schema.defaultSnippets) {
        if (node.items.length === 1) {
          this.collectDefaultSnippets(schema2.schema, separatorAfter, collector, {
            newLineFirst: false,
            indentFirstObject: false,
            shouldIndentWithTab: true
          }, 1);
        } else {
          this.collectDefaultSnippets(schema2.schema, separatorAfter, collector, {
            newLineFirst: false,
            indentFirstObject: true,
            shouldIndentWithTab: false
          }, 1);
        }
      }
    }
  }
  getValueCompletions(schema, doc, node, offset, document, collector, types) {
    let parentKey = null;
    if (node && isScalar4(node)) {
      node = doc.getParent(node);
    }
    if (!node) {
      this.addSchemaValueCompletions(schema.schema, "", collector, types);
      return;
    }
    if (isPair3(node)) {
      const valueNode = node.value;
      if (valueNode && valueNode.range && offset > valueNode.range[0] + valueNode.range[2]) {
        return;
      }
      parentKey = isScalar4(node.key) ? node.key.value.toString() : null;
      node = doc.getParent(node);
    }
    if (node && (parentKey !== null || isSeq3(node))) {
      const separatorAfter = "";
      const matchingSchemas = doc.getMatchingSchemas(schema.schema);
      for (const s of matchingSchemas) {
        if (s.node.internalNode === node && !s.inverted && s.schema) {
          if (s.schema.items) {
            this.collectDefaultSnippets(s.schema, separatorAfter, collector, {
              newLineFirst: false,
              indentFirstObject: false,
              shouldIndentWithTab: false
            });
            if (isSeq3(node) && node.items) {
              if (Array.isArray(s.schema.items)) {
                const index = this.findItemAtOffset(node, document, offset);
                if (index < s.schema.items.length) {
                  this.addSchemaValueCompletions(s.schema.items[index], separatorAfter, collector, types);
                }
              } else if (typeof s.schema.items === "object" && s.schema.items.type === "object") {
                collector.add({
                  kind: this.getSuggestionKind(s.schema.items.type),
                  label: "- (array item)",
                  documentation: `Create an item of an array${s.schema.description === void 0 ? "" : "(" + s.schema.description + ")"}`,
                  insertText: `- ${this.getInsertTextForObject(s.schema.items, separatorAfter, "  ").insertText.trimLeft()}`,
                  insertTextFormat: InsertTextFormat2.Snippet
                });
                this.addSchemaValueCompletions(s.schema.items, separatorAfter, collector, types);
              } else if (typeof s.schema.items === "object" && s.schema.items.anyOf) {
                s.schema.items.anyOf.filter((i) => typeof i === "object").forEach((i, index) => {
                  const insertText = `- ${this.getInsertTextForObject(i, separatorAfter).insertText.trimLeft()}`;
                  const documentation = this.getDocumentationWithMarkdownText(`Create an item of an array${s.schema.description === void 0 ? "" : "(" + s.schema.description + ")"}`, insertText);
                  collector.add({
                    kind: this.getSuggestionKind(i.type),
                    label: "- (array item) " + (index + 1),
                    documentation,
                    insertText,
                    insertTextFormat: InsertTextFormat2.Snippet
                  });
                });
                this.addSchemaValueCompletions(s.schema.items, separatorAfter, collector, types);
              } else {
                this.addSchemaValueCompletions(s.schema.items, separatorAfter, collector, types);
              }
            }
          }
          if (s.schema.properties) {
            const propertySchema = s.schema.properties[parentKey];
            if (propertySchema) {
              this.addSchemaValueCompletions(propertySchema, separatorAfter, collector, types);
            }
          }
        }
      }
      if (types["boolean"]) {
        this.addBooleanValueCompletion(true, separatorAfter, collector);
        this.addBooleanValueCompletion(false, separatorAfter, collector);
      }
      if (types["null"]) {
        this.addNullValueCompletion(separatorAfter, collector);
      }
    }
  }
  getInsertTextForProperty(key, propertySchema, separatorAfter, ident = this.indentation) {
    const propertyText = this.getInsertTextForValue(key, "", "string");
    const resultText = propertyText + ":";
    let value;
    let nValueProposals = 0;
    if (propertySchema) {
      let type = Array.isArray(propertySchema.type) ? propertySchema.type[0] : propertySchema.type;
      if (!type) {
        if (propertySchema.properties) {
          type = "object";
        } else if (propertySchema.items) {
          type = "array";
        }
      }
      if (Array.isArray(propertySchema.defaultSnippets)) {
        if (propertySchema.defaultSnippets.length === 1) {
          const body = propertySchema.defaultSnippets[0].body;
          if (isDefined2(body)) {
            value = this.getInsertTextForSnippetValue(body, "", {
              newLineFirst: true,
              indentFirstObject: false,
              shouldIndentWithTab: false
            }, 1);
            if (!value.startsWith(" ") && !value.startsWith("\n")) {
              value = " " + value;
            }
          }
        }
        nValueProposals += propertySchema.defaultSnippets.length;
      }
      if (propertySchema.enum) {
        if (!value && propertySchema.enum.length === 1) {
          value = " " + this.getInsertTextForGuessedValue(propertySchema.enum[0], "", type);
        }
        nValueProposals += propertySchema.enum.length;
      }
      if (isDefined2(propertySchema.default)) {
        if (!value) {
          value = " " + this.getInsertTextForGuessedValue(propertySchema.default, "", type);
        }
        nValueProposals++;
      }
      if (Array.isArray(propertySchema.examples) && propertySchema.examples.length) {
        if (!value) {
          value = " " + this.getInsertTextForGuessedValue(propertySchema.examples[0], "", type);
        }
        nValueProposals += propertySchema.examples.length;
      }
      if (propertySchema.properties) {
        return `${resultText}
${this.getInsertTextForObject(propertySchema, separatorAfter, ident).insertText}`;
      } else if (propertySchema.items) {
        return `${resultText}
${this.indentation}- ${this.getInsertTextForArray(propertySchema.items, separatorAfter).insertText}`;
      }
      if (nValueProposals === 0) {
        switch (type) {
          case "boolean":
            value = " $1";
            break;
          case "string":
            value = " $1";
            break;
          case "object":
            value = `
${ident}`;
            break;
          case "array":
            value = `
${ident}- `;
            break;
          case "number":
          case "integer":
            value = " ${1:0}";
            break;
          case "null":
            value = " ${1:null}";
            break;
          default:
            return propertyText;
        }
      }
    }
    if (!value || nValueProposals > 1) {
      value = " $1";
    }
    return resultText + value + separatorAfter;
  }
  getInsertTextForObject(schema, separatorAfter, indent = this.indentation, insertIndex = 1) {
    let insertText = "";
    if (!schema.properties) {
      insertText = `${indent}$${insertIndex++}
`;
      return { insertText, insertIndex };
    }
    Object.keys(schema.properties).forEach((key) => {
      const propertySchema = schema.properties[key];
      let type = Array.isArray(propertySchema.type) ? propertySchema.type[0] : propertySchema.type;
      if (!type) {
        if (propertySchema.properties) {
          type = "object";
        }
        if (propertySchema.items) {
          type = "array";
        }
      }
      if (schema.required && schema.required.indexOf(key) > -1) {
        switch (type) {
          case "boolean":
          case "string":
          case "number":
          case "integer":
            insertText += `${indent}${key}: $${insertIndex++}
`;
            break;
          case "array":
            {
              const arrayInsertResult = this.getInsertTextForArray(propertySchema.items, separatorAfter, insertIndex++);
              const arrayInsertLines = arrayInsertResult.insertText.split("\n");
              let arrayTemplate = arrayInsertResult.insertText;
              if (arrayInsertLines.length > 1) {
                for (let index = 1; index < arrayInsertLines.length; index++) {
                  const element = arrayInsertLines[index];
                  arrayInsertLines[index] = `${indent}${this.indentation}  ${element.trimLeft()}`;
                }
                arrayTemplate = arrayInsertLines.join("\n");
              }
              insertIndex = arrayInsertResult.insertIndex;
              insertText += `${indent}${key}:
${indent}${this.indentation}- ${arrayTemplate}
`;
            }
            break;
          case "object":
            {
              const objectInsertResult = this.getInsertTextForObject(propertySchema, separatorAfter, `${indent}${this.indentation}`, insertIndex++);
              insertIndex = objectInsertResult.insertIndex;
              insertText += `${indent}${key}:
${objectInsertResult.insertText}
`;
            }
            break;
        }
      } else if (propertySchema.default !== void 0) {
        switch (type) {
          case "boolean":
          case "number":
          case "integer":
            insertText += `${indent}${key}: \${${insertIndex++}:${propertySchema.default}}
`;
            break;
          case "string":
            insertText += `${indent}${key}: \${${insertIndex++}:${convertToStringValue(propertySchema.default)}}
`;
            break;
          case "array":
          case "object":
            break;
        }
      }
    });
    if (insertText.trim().length === 0) {
      insertText = `${indent}$${insertIndex++}
`;
    }
    insertText = insertText.trimRight() + separatorAfter;
    return { insertText, insertIndex };
  }
  getInsertTextForArray(schema, separatorAfter, insertIndex = 1) {
    let insertText = "";
    if (!schema) {
      insertText = `$${insertIndex++}`;
      return { insertText, insertIndex };
    }
    let type = Array.isArray(schema.type) ? schema.type[0] : schema.type;
    if (!type) {
      if (schema.properties) {
        type = "object";
      }
      if (schema.items) {
        type = "array";
      }
    }
    switch (schema.type) {
      case "boolean":
        insertText = `\${${insertIndex++}:false}`;
        break;
      case "number":
      case "integer":
        insertText = `\${${insertIndex++}:0}`;
        break;
      case "string":
        insertText = `\${${insertIndex++}:""}`;
        break;
      case "object":
        {
          const objectInsertResult = this.getInsertTextForObject(schema, separatorAfter, `${this.indentation}  `, insertIndex++);
          insertText = objectInsertResult.insertText.trimLeft();
          insertIndex = objectInsertResult.insertIndex;
        }
        break;
    }
    return { insertText, insertIndex };
  }
  getInsertTextForGuessedValue(value, separatorAfter, type) {
    switch (typeof value) {
      case "object":
        if (value === null) {
          return "${1:null}" + separatorAfter;
        }
        return this.getInsertTextForValue(value, separatorAfter, type);
      case "string": {
        let snippetValue = JSON.stringify(value);
        snippetValue = snippetValue.substr(1, snippetValue.length - 2);
        snippetValue = this.getInsertTextForPlainText(snippetValue);
        if (type === "string") {
          snippetValue = convertToStringValue(snippetValue);
        }
        return "${1:" + snippetValue + "}" + separatorAfter;
      }
      case "number":
      case "boolean":
        return "${1:" + value + "}" + separatorAfter;
    }
    return this.getInsertTextForValue(value, separatorAfter, type);
  }
  getInsertTextForPlainText(text) {
    return text.replace(/[\\$}]/g, "\\$&");
  }
  getInsertTextForValue(value, separatorAfter, type) {
    if (value === null) {
      value = "null";
    }
    switch (typeof value) {
      case "object": {
        const indent = this.indentation;
        return this.getInsertTemplateForValue(value, indent, { index: 1 }, separatorAfter);
      }
    }
    type = Array.isArray(type) ? type[0] : type;
    if (type === "string") {
      value = convertToStringValue(value);
    }
    return this.getInsertTextForPlainText(value + separatorAfter);
  }
  getInsertTemplateForValue(value, indent, navOrder, separatorAfter) {
    if (Array.isArray(value)) {
      let insertText = "\n";
      for (const arrValue of value) {
        insertText += `${indent}- \${${navOrder.index++}:${arrValue}}
`;
      }
      return insertText;
    } else if (typeof value === "object") {
      let insertText = "\n";
      for (const key in value) {
        if (Object.prototype.hasOwnProperty.call(value, key)) {
          const element = value[key];
          insertText += `${indent}\${${navOrder.index++}:${key}}:`;
          let valueTemplate;
          if (typeof element === "object") {
            valueTemplate = `${this.getInsertTemplateForValue(element, indent + this.indentation, navOrder, separatorAfter)}`;
          } else {
            valueTemplate = ` \${${navOrder.index++}:${this.getInsertTextForPlainText(element + separatorAfter)}}
`;
          }
          insertText += `${valueTemplate}`;
        }
      }
      return insertText;
    }
    return this.getInsertTextForPlainText(value + separatorAfter);
  }
  addSchemaValueCompletions(schema, separatorAfter, collector, types) {
    if (typeof schema === "object") {
      this.addEnumValueCompletions(schema, separatorAfter, collector);
      this.addDefaultValueCompletions(schema, separatorAfter, collector);
      this.collectTypes(schema, types);
      if (Array.isArray(schema.allOf)) {
        schema.allOf.forEach((s) => {
          return this.addSchemaValueCompletions(s, separatorAfter, collector, types);
        });
      }
      if (Array.isArray(schema.anyOf)) {
        schema.anyOf.forEach((s) => {
          return this.addSchemaValueCompletions(s, separatorAfter, collector, types);
        });
      }
      if (Array.isArray(schema.oneOf)) {
        schema.oneOf.forEach((s) => {
          return this.addSchemaValueCompletions(s, separatorAfter, collector, types);
        });
      }
    }
  }
  collectTypes(schema, types) {
    if (Array.isArray(schema.enum) || isDefined2(schema.const)) {
      return;
    }
    const type = schema.type;
    if (Array.isArray(type)) {
      type.forEach(function(t) {
        return types[t] = true;
      });
    } else if (type) {
      types[type] = true;
    }
  }
  addDefaultValueCompletions(schema, separatorAfter, collector, arrayDepth = 0) {
    let hasProposals = false;
    if (isDefined2(schema.default)) {
      let type = schema.type;
      let value = schema.default;
      for (let i = arrayDepth; i > 0; i--) {
        value = [value];
        type = "array";
      }
      let label;
      if (typeof value == "object") {
        label = "Default value";
      } else {
        label = value.toString().replace(doubleQuotesEscapeRegExp, '"');
      }
      collector.add({
        kind: this.getSuggestionKind(type),
        label,
        insertText: this.getInsertTextForValue(value, separatorAfter, type),
        insertTextFormat: InsertTextFormat2.Snippet,
        detail: localize9("json.suggest.default", "Default value")
      });
      hasProposals = true;
    }
    if (Array.isArray(schema.examples)) {
      schema.examples.forEach((example) => {
        let type = schema.type;
        let value = example;
        for (let i = arrayDepth; i > 0; i--) {
          value = [value];
          type = "array";
        }
        collector.add({
          kind: this.getSuggestionKind(type),
          label: this.getLabelForValue(value),
          insertText: this.getInsertTextForValue(value, separatorAfter, type),
          insertTextFormat: InsertTextFormat2.Snippet
        });
        hasProposals = true;
      });
    }
    this.collectDefaultSnippets(schema, separatorAfter, collector, {
      newLineFirst: true,
      indentFirstObject: true,
      shouldIndentWithTab: true
    });
    if (!hasProposals && typeof schema.items === "object" && !Array.isArray(schema.items)) {
      this.addDefaultValueCompletions(schema.items, separatorAfter, collector, arrayDepth + 1);
    }
  }
  addEnumValueCompletions(schema, separatorAfter, collector) {
    if (isDefined2(schema.const)) {
      collector.add({
        kind: this.getSuggestionKind(schema.type),
        label: this.getLabelForValue(schema.const),
        insertText: this.getInsertTextForValue(schema.const, separatorAfter, void 0),
        insertTextFormat: InsertTextFormat2.Snippet,
        documentation: this.fromMarkup(schema.markdownDescription) || schema.description
      });
    }
    if (Array.isArray(schema.enum)) {
      for (let i = 0, length = schema.enum.length; i < length; i++) {
        const enm = schema.enum[i];
        let documentation = this.fromMarkup(schema.markdownDescription) || schema.description;
        if (schema.markdownEnumDescriptions && i < schema.markdownEnumDescriptions.length && this.doesSupportMarkdown()) {
          documentation = this.fromMarkup(schema.markdownEnumDescriptions[i]);
        } else if (schema.enumDescriptions && i < schema.enumDescriptions.length) {
          documentation = schema.enumDescriptions[i];
        }
        collector.add({
          kind: this.getSuggestionKind(schema.type),
          label: this.getLabelForValue(enm),
          insertText: this.getInsertTextForValue(enm, separatorAfter, void 0),
          insertTextFormat: InsertTextFormat2.Snippet,
          documentation
        });
      }
    }
  }
  getLabelForValue(value) {
    if (value === null) {
      return "null";
    }
    if (Array.isArray(value)) {
      return JSON.stringify(value);
    }
    return value;
  }
  collectDefaultSnippets(schema, separatorAfter, collector, settings, arrayDepth = 0) {
    if (Array.isArray(schema.defaultSnippets)) {
      for (const s of schema.defaultSnippets) {
        let type = schema.type;
        let value = s.body;
        let label = s.label;
        let insertText;
        let filterText;
        if (isDefined2(value)) {
          const type2 = s.type || schema.type;
          if (arrayDepth === 0 && type2 === "array") {
            const fixedObj = {};
            Object.keys(value).forEach((val, index) => {
              if (index === 0 && !val.startsWith("-")) {
                fixedObj[`- ${val}`] = value[val];
              } else {
                fixedObj[`  ${val}`] = value[val];
              }
            });
            value = fixedObj;
          }
          insertText = this.getInsertTextForSnippetValue(value, separatorAfter, settings);
          label = label || this.getLabelForSnippetValue(value);
        } else if (typeof s.bodyText === "string") {
          let prefix = "", suffix = "", indent = "";
          for (let i = arrayDepth; i > 0; i--) {
            prefix = prefix + indent + "[\n";
            suffix = suffix + "\n" + indent + "]";
            indent += this.indentation;
            type = "array";
          }
          insertText = prefix + indent + s.bodyText.split("\n").join("\n" + indent) + suffix + separatorAfter;
          label = label || insertText;
          filterText = insertText.replace(/[\n]/g, "");
        }
        collector.add({
          kind: s.suggestionKind || this.getSuggestionKind(type),
          label,
          documentation: this.fromMarkup(s.markdownDescription) || s.description,
          insertText,
          insertTextFormat: InsertTextFormat2.Snippet,
          filterText
        });
      }
    }
  }
  getInsertTextForSnippetValue(value, separatorAfter, settings, depth) {
    const replacer = (value2) => {
      if (typeof value2 === "string") {
        if (value2[0] === "^") {
          return value2.substr(1);
        }
        if (value2 === "true" || value2 === "false") {
          return `"${value2}"`;
        }
      }
      return value2;
    };
    return stringifyObject2(value, "", replacer, settings, depth) + separatorAfter;
  }
  addBooleanValueCompletion(value, separatorAfter, collector) {
    collector.add({
      kind: this.getSuggestionKind("boolean"),
      label: value ? "true" : "false",
      insertText: this.getInsertTextForValue(value, separatorAfter, "boolean"),
      insertTextFormat: InsertTextFormat2.Snippet,
      documentation: ""
    });
  }
  addNullValueCompletion(separatorAfter, collector) {
    collector.add({
      kind: this.getSuggestionKind("null"),
      label: "null",
      insertText: "null" + separatorAfter,
      insertTextFormat: InsertTextFormat2.Snippet,
      documentation: ""
    });
  }
  getLabelForSnippetValue(value) {
    const label = JSON.stringify(value);
    return label.replace(/\$\{\d+:([^}]+)\}|\$\d+/g, "$1");
  }
  getCustomTagValueCompletions(collector) {
    const validCustomTags = filterInvalidCustomTags(this.customTags);
    validCustomTags.forEach((validTag) => {
      const label = validTag.split(" ")[0];
      this.addCustomTagValueCompletion(collector, " ", label);
    });
  }
  addCustomTagValueCompletion(collector, separatorAfter, label) {
    collector.add({
      kind: this.getSuggestionKind("string"),
      label,
      insertText: label + separatorAfter,
      insertTextFormat: InsertTextFormat2.Snippet,
      documentation: ""
    });
  }
  getDocumentationWithMarkdownText(documentation, insertText) {
    let res = documentation;
    if (this.doesSupportMarkdown()) {
      insertText = insertText.replace(/\${[0-9]+[:|](.*)}/g, (s, arg) => {
        return arg;
      }).replace(/\$([0-9]+)/g, "");
      res = this.fromMarkup(`${documentation}
 \`\`\`
${insertText}
\`\`\``);
    }
    return res;
  }
  getSuggestionKind(type) {
    if (Array.isArray(type)) {
      const array = type;
      type = array.length > 0 ? array[0] : null;
    }
    if (!type) {
      return CompletionItemKind2.Value;
    }
    switch (type) {
      case "string":
        return CompletionItemKind2.Value;
      case "object":
        return CompletionItemKind2.Module;
      case "property":
        return CompletionItemKind2.Property;
      default:
        return CompletionItemKind2.Value;
    }
  }
  getCurrentWord(doc, offset) {
    let i = offset - 1;
    const text = doc.getText();
    while (i >= 0 && ' 	\n\r\v":{[,]}'.indexOf(text.charAt(i)) === -1) {
      i--;
    }
    return text.substring(i + 1, offset);
  }
  fromMarkup(markupString) {
    if (markupString && this.doesSupportMarkdown()) {
      return {
        kind: MarkupKind2.Markdown,
        value: markupString
      };
    }
    return void 0;
  }
  doesSupportMarkdown() {
    if (this.supportsMarkdown === void 0) {
      const completion = this.clientCapabilities.textDocument && this.clientCapabilities.textDocument.completion;
      this.supportsMarkdown = completion && completion.completionItem && Array.isArray(completion.completionItem.documentationFormat) && completion.completionItem.documentationFormat.indexOf(MarkupKind2.Markdown) !== -1;
    }
    return this.supportsMarkdown;
  }
  findItemAtOffset(seqNode, doc, offset) {
    for (let i = seqNode.items.length - 1; i >= 0; i--) {
      const node = seqNode.items[i];
      if (isNode2(node)) {
        if (node.range) {
          if (offset > node.range[1]) {
            return i;
          } else if (offset >= node.range[0]) {
            return i;
          }
        }
      }
    }
    return 0;
  }
};
var isNumberExp = /^\d+$/;
function convertToStringValue(value) {
  if (value.length === 0) {
    return value;
  }
  if (value === "true" || value === "false" || value === "null" || isNumberExp.test(value)) {
    return `"${value}"`;
  }
  if (value.indexOf('"') !== -1) {
    value = value.replace(doubleQuotesEscapeRegExp, '"');
  }
  let doQuote = value.charAt(0) === "@";
  if (!doQuote) {
    let idx = value.indexOf(":", 0);
    for (; idx > 0 && idx < value.length; idx = value.indexOf(":", idx + 1)) {
      if (idx === value.length - 1) {
        doQuote = true;
        break;
      }
      const nextChar = value.charAt(idx + 1);
      if (nextChar === "	" || nextChar === " ") {
        doQuote = true;
        break;
      }
    }
  }
  if (doQuote) {
    value = `"${value}"`;
  }
  return value;
}

// node_modules/yaml-language-server/lib/esm/languageservice/services/yamlDefinition.js
import { LocationLink, Range as Range11 } from "vscode-languageserver-types";
import { isAlias as isAlias2 } from "yaml";
function getDefinition(document, params) {
  try {
    const yamlDocument = yamlDocumentsCache.getYamlDocument(document);
    const offset = document.offsetAt(params.position);
    const currentDoc = matchOffsetToDocument(offset, yamlDocument);
    if (currentDoc) {
      const [node] = currentDoc.getNodeFromPosition(offset, new TextBuffer(document));
      if (node && isAlias2(node)) {
        const defNode = node.resolve(currentDoc.internalDocument);
        if (defNode && defNode.range) {
          const targetRange = Range11.create(document.positionAt(defNode.range[0]), document.positionAt(defNode.range[2]));
          const selectionRange = Range11.create(document.positionAt(defNode.range[0]), document.positionAt(defNode.range[1]));
          return [LocationLink.create(document.uri, targetRange, selectionRange)];
        }
      }
    }
  } catch (err) {
    this.telemetry.sendError("yaml.definition.error", { error: err });
  }
  return void 0;
}

// node_modules/yaml-language-server/lib/esm/languageservice/yamlLanguageService.js
var SchemaPriority;
(function(SchemaPriority2) {
  SchemaPriority2[SchemaPriority2["SchemaStore"] = 1] = "SchemaStore";
  SchemaPriority2[SchemaPriority2["SchemaAssociation"] = 2] = "SchemaAssociation";
  SchemaPriority2[SchemaPriority2["Settings"] = 3] = "Settings";
  SchemaPriority2[SchemaPriority2["Modeline"] = 4] = "Modeline";
})(SchemaPriority || (SchemaPriority = {}));
function getLanguageService(schemaRequestService2, workspaceContext, connection, telemetry, clientCapabilities) {
  const schemaService = new YAMLSchemaService(schemaRequestService2, workspaceContext);
  const completer = new YamlCompletion(schemaService, clientCapabilities, yamlDocumentsCache, telemetry);
  const hover = new YAMLHover(schemaService, telemetry);
  const yamlDocumentSymbols = new YAMLDocumentSymbols(schemaService, telemetry);
  const yamlValidation = new YAMLValidation(schemaService);
  const formatter = new YAMLFormatter();
  const yamlCodeActions = new YamlCodeActions(clientCapabilities);
  const yamlCodeLens = new YamlCodeLens(schemaService, telemetry);
  registerCommands(commandExecutor, connection);
  return {
    configure: (settings) => {
      schemaService.clearExternalSchemas();
      if (settings.schemas) {
        schemaService.schemaPriorityMapping = new Map();
        settings.schemas.forEach((settings2) => {
          const currPriority = settings2.priority ? settings2.priority : 0;
          schemaService.addSchemaPriority(settings2.uri, currPriority);
          schemaService.registerExternalSchema(settings2.uri, settings2.fileMatch, settings2.schema, settings2.name, settings2.description);
        });
      }
      yamlValidation.configure(settings);
      hover.configure(settings);
      completer.configure(settings);
      formatter.configure(settings);
      yamlCodeActions.configure(settings);
    },
    registerCustomSchemaProvider: (schemaProvider) => {
      schemaService.registerCustomSchemaProvider(schemaProvider);
    },
    findLinks: findLinks2,
    doComplete: completer.doComplete.bind(completer),
    doValidation: yamlValidation.doValidation.bind(yamlValidation),
    doHover: hover.doHover.bind(hover),
    findDocumentSymbols: yamlDocumentSymbols.findDocumentSymbols.bind(yamlDocumentSymbols),
    findDocumentSymbols2: yamlDocumentSymbols.findHierarchicalDocumentSymbols.bind(yamlDocumentSymbols),
    doDefinition: getDefinition.bind(getDefinition),
    resetSchema: (uri) => {
      return schemaService.onResourceChange(uri);
    },
    doFormat: formatter.format.bind(formatter),
    doDocumentOnTypeFormatting,
    addSchema: (schemaID, schema) => {
      return schemaService.saveSchema(schemaID, schema);
    },
    deleteSchema: (schemaID) => {
      return schemaService.deleteSchema(schemaID);
    },
    modifySchemaContent: (schemaAdditions) => {
      return schemaService.addContent(schemaAdditions);
    },
    deleteSchemaContent: (schemaDeletions) => {
      return schemaService.deleteContent(schemaDeletions);
    },
    deleteSchemasWhole: (schemaDeletions) => {
      return schemaService.deleteSchemas(schemaDeletions);
    },
    getFoldingRanges: getFoldingRanges2,
    getCodeAction: (document, params) => {
      return yamlCodeActions.getCodeAction(document, params);
    },
    getCodeLens: (document, params) => {
      return yamlCodeLens.getCodeLens(document, params);
    },
    resolveCodeLens: (param) => yamlCodeLens.resolveCodeLens(param)
  };
}

// src/constants.ts
var languageId = "yaml";

// src/yamlWorker.ts
async function schemaRequestService(uri) {
  const response = await fetch(uri);
  if (response.ok) {
    return response.text();
  }
  throw new Error(`Schema request failed for ${uri}`);
}
function createYAMLWorker(ctx, { enableSchemaRequest, languageSettings }) {
  const languageService = getLanguageService(enableSchemaRequest ? schemaRequestService : null, null, null, null);
  languageService.configure(languageSettings);
  const getTextDocument = (uri) => {
    const models = ctx.getMirrorModels();
    for (const model of models) {
      if (String(model.uri) === uri) {
        return TextDocument2.create(uri, languageId, model.version, model.getValue());
      }
    }
    return null;
  };
  return {
    doValidation(uri) {
      const document = getTextDocument(uri);
      if (document) {
        return languageService.doValidation(document, languageSettings.isKubernetes);
      }
      return [];
    },
    doComplete(uri, position) {
      const document = getTextDocument(uri);
      return languageService.doComplete(document, position, languageSettings.isKubernetes);
    },
    doDefinition(uri, position) {
      const document = getTextDocument(uri);
      return languageService.doDefinition(document, { position, textDocument: { uri } });
    },
    doHover(uri, position) {
      const document = getTextDocument(uri);
      return languageService.doHover(document, position);
    },
    format(uri, options) {
      const document = getTextDocument(uri);
      return languageService.doFormat(document, options);
    },
    resetSchema(uri) {
      return languageService.resetSchema(uri);
    },
    findDocumentSymbols(uri) {
      const document = getTextDocument(uri);
      return languageService.findDocumentSymbols2(document, {});
    },
    findLinks(uri) {
      const document = getTextDocument(uri);
      return Promise.resolve(languageService.findLinks(document));
    }
  };
}

// src/yaml.worker.ts
self.onmessage = () => {
  initialize((ctx, createData) => Object.create(createYAMLWorker(ctx, createData)));
};
//# sourceMappingURL=yaml.worker.js.map
