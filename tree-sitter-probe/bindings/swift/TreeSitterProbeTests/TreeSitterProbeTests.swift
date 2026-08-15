import XCTest
import SwiftTreeSitter
import TreeSitterProbe

final class TreeSitterProbeTests: XCTestCase {
    func testCanLoadGrammar() throws {
        let parser = Parser()
        let language = Language(language: tree_sitter_probe())
        XCTAssertNoThrow(try parser.setLanguage(language),
                         "Error loading Probe grammar")
    }
}
