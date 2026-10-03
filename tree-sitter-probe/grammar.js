/**
 * @file Probe is a language for describing API requests
 * @author Charuka Karunanayake <charukakarunanayake@gmail.com>
 * @license Apache-2.0
 */

/// <reference types="tree-sitter-cli/dsl" />

// @ts-check

export default grammar({
  name: "probe",

  rules: {
    source_file: $ => seq(
      repeat($._newline),
      $.request,
      repeat($._newline)
    ),
    _newline: _ => /\r?\n/,
    request_type: _ => choice("GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS", "CONNECT", "TRACE"),
    request_target: _ => /[^\s]+/,
    http_version: _ => /HTTP\/\d+(\.\d+)?/,
    request_line: $ => seq(
      field('type', $.request_type),
      field('target', $.request_target),
      optional(field('version', $.http_version))
    ),
    request: $ => seq(
      field('request_line', $.request_line),
      $._newline
    )
  }
});
