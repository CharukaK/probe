/**
 * @file Parser for Probe, a text-based API testing DSL built on HTTP syntax
 *       (RFC 9110/9112) with request-file extensions (multi-request files,
 *       multipart shorthand, and more)
 * @author Charuka Karunanayake <charukakarunanayake@gmail.com>
 * @license MIT
 */

/// <reference types="tree-sitter-cli/dsl" />
// @ts-check

module.exports = grammar({
  name: "probe",
  extras: _ => [],
  externals: $ => [$._octet_body],
  rules: {
    // --- Entry point ---
    source_file: $ => seq(
      repeat($._line_ending),
      repeat(seq($.let_directive, repeat($._line_ending))),
      optional($.seprator),
      $.request_block,
      repeat(seq(
        repeat($._line_ending),
        $.seprator,
        repeat($._line_ending),
        $.request_block
      )),
      repeat($._line_ending)
    ),

    // --- Low-level primitives ---
    _line_ending: _ => choice('\r\n', '\n'),
    _wsp: _ => /[ \t]/,

    identifier: _ => /[a-zA-Z][a-zA-Z0-9_]*/,
    _octet: _ => /[\s\S]/,
    digit: _ => /[0-9]/,
    octet_body: $ => $._octet_body,

    _target_run: _ => /[!-z|-~]+/,
    _value_run: _ => /[!-z|-~ \t]+/,
    comparator_operator: _ => choice('==', '<', '<=', '=>', '>', 'contains', 'matches'),

    _string_content: _ => token.immediate(prec(1, /[^"\\]+/)),
    _escape_sequence: _ => token.immediate(seq(
      '\\',
      choice(/["\\/bfnrt]/, seq('u', /[0-9a-fA-F]{4}/))
    )),
    string: $ => seq(
      '"',
      repeat(choice($._string_content, $._escape_sequence)),
      '"'
    ),
    number: _ => /-?(0|[1-9][0-9]*)(\.[0-9]+)?([eE][+-]?[0-9]+)?/,
    argument: $ => choice($.value_reference, $.string, $.number),
    arguments: $ => seq(
      field('argument', $.argument),
      repeat(seq(
        repeat($._wsp),
        ',',
        repeat($._wsp),
        field('argument', $.argument)
      ))
    ),

    // --- Lexical tokens ---
    seprator: $ => seq(
      '###',
      optional(seq(repeat1($._wsp), field('name', $.request_name))),
      $._line_ending
    ),
    request_name: _ => /[ -~]+/,
    method: _ => /[!#$%&'*+\-.^_`|~0-9A-Za-z]+/,
    request_target: $ => repeat1(choice($.interpolation, $._target_run, '{')), // capture the token
    field_name: _ => /[!#$%&'*+\-.^_`|~0-9A-Za-z]+/,
    field_value: $ => repeat1(choice($.interpolation, $._value_run, '{')),
    field_value_seperator: _ => ':',
    multipart_part: $ => seq(choice('@field', '@file'), /[^\r\n]*/, $._line_ending),
    multipart_body: $ => repeat1($.multipart_part),
    message_body: $ => choice($.multipart_body, $.octet_body),
    // A chain of accessors walking into a value, root-first. This is the
    // one building block for every dotted/indexed path in the language:
    // {{alias.name}} today, and later @assert's `body.x.y[0]` target
    // (docs/language-spec.md §7.1) reuses `accessor` unchanged, rooted at
    // the literal `body` instead of an `identifier`, plus an
    // `index_access` alternative added to the choice below.
    member_access: $ => seq(
      '.',
      field('name', $.identifier),
    ),
    index_access: $ => seq(
      '[',
      field('index', choice($.number, $.string)),
      ']'
    ),
    accessor: $ => choice($.member_access, $.index_access),
    value_reference: $ => seq(
      field('root', $.identifier),
      repeat(field('accessor', $.accessor)),
    ),
    header_target: $ => seq('headers', '.', field('name', $.field_name)),
    body_target: $ => seq('body', repeat(field('accessor', $.accessor))),
    assert_target: $ => choice('status', 'duration', $.header_target, $.body_target),

    util_reference: $ => seq(
      field('util_name', $.identifier),
      '(',
      optional($.arguments),
      ')',
    ),
    interpolation_body: $ => choice($.value_reference, $.util_reference),
    interpolation: $ => seq(
      '{{',
      repeat($._wsp),
      field('body', $.interpolation_body),
      repeat($._wsp),
      '}}'
    ),
    json_literal: $ => choice('null', 'true', 'false', $.number, $.string),
    let_value: $ => choice($.json_literal, $.interpolation),
    save_directive: $ => seq(
      '@save',
      $._wsp,
      field('name', $.identifier),
      optional($._wsp),
      '=',
      optional($._wsp),
      field('target', $.assert_target),
      $._line_ending
    ),
    let_directive: $ => seq(
      '@let',
      $._wsp,
      field('name', $.identifier),
      optional($._wsp),
      '=',
      optional($._wsp),
      field('target', $.let_value),
      $._line_ending
    ),
    assert_directive: $ => seq(
      '@assert',
      $._wsp,
      field('target', $.assert_target),
      optional($._wsp),
      field('comparator', $.comparator_operator),
      optional($._wsp),
      field('target', $.let_value),
      $._line_ending
    ),
    directive_line: $ => choice($.save_directive, $.let_directive, $.assert_directive),

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

    request_block: $ => seq(
      $.request_line,
      repeat($.field_line),
      $._line_ending,
      optional($.message_body),
      repeat($.directive_line)
    )
  }
});
