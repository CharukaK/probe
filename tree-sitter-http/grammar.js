/**
 * @file Parser for HTTP syntax specified in https://datatracker.ietf.org/doc/html/rfc9110
 * @author Charuka Karunanayake <charukakarunanayake@gmail.com>
 * @license MIT
 */

/// <reference types="tree-sitter-cli/dsl" />
// @ts-check

module.exports = grammar({
  name: "http",

  rules: {
    // TODO: add the actual grammar rules
    source_file: $ => "hello"
  }
});
