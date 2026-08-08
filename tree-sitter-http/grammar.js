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
    source_file: $ => $.request,
    _line_ending: _ => choice('\r\n', '\n'),
    _wsp: _ => /[ \t]/,
    digit: _ => /[0-9]/,
    http_version: $ => seq(
      'HTTP/',
      field('major', $.digit),
      '.',
      field('minor', $.digit)
    ),
    method: _ => /[!#$%&'*+\-.^_`|~0-9A-Za-z]+/,
    field_name: _ => /[!#$%&'*+\-.^_`|~0-9A-Za-z]+/,
    field_value: _ => /[!-~]+(?:[ \t]+[!-~]+)*/,
    field_value_seperator: _ => ':',
    request_target: _ => /[!-~]+/, // capture the token
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
    request: $ => seq(
      $.request_line,
      optional(repeat($.field_line))
    )
  }
});
