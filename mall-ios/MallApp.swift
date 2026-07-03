import SwiftUI

@main
struct MallApp: App {
    init() {
        MainActor.assumeIsolated {
            AppSession.shared.bootstrap()
        }
    }

    var body: some Scene {
        WindowGroup {
            ContentView()
        }
    }
}
