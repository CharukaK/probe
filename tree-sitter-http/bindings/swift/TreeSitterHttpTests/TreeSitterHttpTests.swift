import XCTest
import SwiftTreeSitter
import TreeSitterHttp

final class TreeSitterHttpTests: XCTestCase {
    func testCanLoadGrammar() throws {
        let parser = Parser()
        let language = Language(language: tree_sitter_http())
        XCTAssertNoThrow(try parser.setLanguage(language),
                         "Error loading HTTP grammar")
    }
}
