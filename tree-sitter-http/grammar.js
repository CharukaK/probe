/**
 * @file Parser for HTTP syntax specified in https://datatracker.ietf.org/doc/html/rfc9110
 * @author Charuka Karunanayake <charukakarunanayake@gmail.com>
 * @license MIT
 */

/// <reference types="tree-sitter-cli/dsl" />
// @ts-check

module.exports = grammar({
  name: "http",
  extras: _ => [],
  rules: {
    // --- Entry point ---
    source_file: $ => $.request_block,

    // --- Low-level primitives ---
    _line_ending: _ => choice('\r\n', '\n'),
    _wsp: _ => /[ \t]/,
    _octet: _ => /[\s\S]/,
    digit: _ => /[0-9]/,

    // --- Lexical tokens ---
    method: _ => /[!#$%&'*+\-.^_`|~0-9A-Za-z]+/,
    request_target: _ => /[!-~]+/, // capture the token
    field_name: _ => /[!#$%&'*+\-.^_`|~0-9A-Za-z]+/,
    field_value: _ => /[!-~]+(?:[ \t]+[!-~]+)*/,
    field_value_seperator: _ => ':',

    http_version: $ => seq(
      'HTTP/',
      field('major', $.digit),
      '.',
      field('minor', $.digit)
    ),

    // --- Structural / composite rules ---
    request_line: $ => seq(
      field('method', $.method),
      $._wsp,
      field('target', $.request_target),
      optional(
        seq(
          $._wsp,
          field('version', $.http_version)
        )
      ),
      $._line_ending
    ),

    field_line: $ => seq(
      field('name', $.field_name),
      $.field_value_seperator,
      repeat($._wsp),
      field('value', $.field_value),
      repeat($._wsp),
      $._line_ending
    ),

    request_body: $ => seq(
      choice($._octet),
      $._line_ending
    ),

    request_block: $ => seq(
      $.request_line,
      optional(repeat($.field_line)),
      optional(
        seq(
          $._line_ending,
          $.request_body,
        )
      ),
    )
  }
});
