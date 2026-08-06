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
    _line_ending: _ => choice('\r\n', '\n'),
    _WS: _ => ' ',
    http_version: _ => seq(
      'HTTP/',
      field('major', /[0-9]/),
      '.',
      field('minor', /[0-9]/)
    ),
    method: _ => choice(
      "GET",
      "POST",
      "PUT",
      "DELETE",
      "PATCH",
      "HEAD",
      "OPTIONS",
      "CONNECT",
      "TRACE"
    ),
    request_target: _ => /[!-~]+/, // capture the token
    request_line: $ => seq(
      field('method', $.method),
      $._WS,
      field('target', $.request_target),
      optional(
        seq(
          $._WS,
          field('version', $.http_version)
        )
      ),
      $._line_ending
    ),
    source_file: _ => "hello",
  }
});
