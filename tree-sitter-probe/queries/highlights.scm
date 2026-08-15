; request line
(request_line method: (method) @keyword)
(request_line target: (request_target) @string.special.url)

"HTTP/" @constant.builtin
(digit) @number
"." @punctuation.delimiter

; Header fields
(field_line name: (field_name) @property)
(field_line value: (field_value) @string)
(field_value_seperator) @punctuation.delimiter

; Multipart body markers
"@field" @attribute
"@file" @attribute

; Raw body content
(octet_body) @string

