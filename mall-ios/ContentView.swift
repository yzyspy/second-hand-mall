import SwiftUI

struct ContentView: View {
    var body: some View {
        TabView {
            HomeView()
                .tabItem {
                    Label("首页", systemImage: "house")
                }
            PublishView()
                .tabItem {
                    Label("发布", systemImage: "plus.circle")
                }
            ProfileView()
                .tabItem {
                    Label("我的", systemImage: "person")
                }
        }
    }
}
