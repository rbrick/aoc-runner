// monaco is bundled with the app instead of loaded from a CDN, so the runner works offline
import { loader } from "@monaco-editor/react";
import "monaco-editor/esm/vs/editor/edcore.main";
import * as monaco from "monaco-editor/esm/vs/editor/editor.api";

import "monaco-editor/esm/vs/basic-languages/go/go.contribution";
import "monaco-editor/esm/vs/basic-languages/java/java.contribution";
import "monaco-editor/esm/vs/basic-languages/python/python.contribution";
import "monaco-editor/esm/vs/basic-languages/rust/rust.contribution";
import "monaco-editor/esm/vs/basic-languages/typescript/typescript.contribution";
import "monaco-editor/esm/vs/language/typescript/monaco.contribution";

import EditorWorker from "monaco-editor/esm/vs/editor/editor.worker?worker";
import TypeScriptWorker from "monaco-editor/esm/vs/language/typescript/ts.worker?worker";

self.MonacoEnvironment = {
  getWorker(_, label) {
    if (label === "typescript" || label === "javascript") {
      return new TypeScriptWorker();
    }

    return new EditorWorker();
  },
};

// monaco doesn't ship odin, this covers enough for highlighting
monaco.languages.register({ id: "odin" });
monaco.languages.setLanguageConfiguration("odin", {
  comments: { lineComment: "//", blockComment: ["/*", "*/"] },
  brackets: [
    ["{", "}"],
    ["[", "]"],
    ["(", ")"],
  ],
  autoClosingPairs: [
    { open: "{", close: "}" },
    { open: "[", close: "]" },
    { open: "(", close: ")" },
    { open: '"', close: '"', notIn: ["string"] },
    { open: "`", close: "`", notIn: ["string"] },
  ],
});
monaco.languages.setMonarchTokensProvider("odin", {
  keywords: [
    "package", "import", "foreign", "proc", "struct", "enum", "union", "bit_set", "map", "dynamic",
    "for", "in", "not_in", "if", "else", "when", "switch", "case", "return", "defer", "break",
    "continue", "fallthrough", "using", "distinct", "cast", "transmute", "auto_cast", "context",
    "or_else", "or_return", "or_break", "or_continue", "where", "matrix",
  ],
  constants: ["true", "false", "nil"],
  types: [
    "int", "uint", "i8", "i16", "i32", "i64", "i128", "u8", "u16", "u32", "u64", "u128",
    "f16", "f32", "f64", "bool", "b8", "b16", "b32", "b64", "byte", "rune", "string", "cstring",
    "rawptr", "uintptr", "any", "typeid",
  ],
  tokenizer: {
    root: [
      [/#[a-z_]+/, "annotation"],
      [/@\(?[a-z_]*/, "annotation"],
      [
        /[a-zA-Z_]\w*/,
        { cases: { "@keywords": "keyword", "@constants": "constant", "@types": "type", "@default": "identifier" } },
      ],
      [/\/\/.*$/, "comment"],
      [/\/\*/, "comment", "@comment"],
      [/"([^"\\]|\\.)*"/, "string"],
      [/`[^`]*`/, "string"],
      [/'([^'\\]|\\.)*'/, "string"],
      [/0[xX][0-9a-fA-F_]+|0[bB][01_]+|\d[\d_]*(\.\d[\d_]*)?([eE][+-]?\d+)?/, "number"],
      [/[{}()[\]]/, "@brackets"],
      [/[:=<>!~?&|+\-*/^%]+/, "operator"],
    ],
    comment: [
      [/[^/*]+/, "comment"],
      [/\*\//, "comment", "@pop"],
      [/[/*]/, "comment"],
    ],
  },
} as monaco.languages.IMonarchLanguage);

// match the app's own surfaces instead of monaco's default themes
monaco.editor.defineTheme("aoc-runner-light", {
  base: "vs",
  inherit: true,
  rules: [],
  colors: {
    "editor.background": "#f1f3f6",
    "editorLineNumber.foreground": "#98a2b3",
    "editorLineNumber.activeForeground": "#3b5bdb",
    "editor.lineHighlightBackground": "#e8edff80",
  },
});

monaco.editor.defineTheme("aoc-runner-dark", {
  base: "vs-dark",
  inherit: true,
  rules: [],
  colors: {
    "editor.background": "#11141a",
    "editorLineNumber.foreground": "#4b5563",
    "editorLineNumber.activeForeground": "#7b93ff",
    "editor.lineHighlightBackground": "#232b4880",
  },
});

loader.config({ monaco });
