import SwiftUI

@main
struct MallApp: App {
    init() {
        AppSession.shared.bootstrap()
    }

    var body: some Scene {
        WindowGroup {
            ContentView()
        }
    }
}
