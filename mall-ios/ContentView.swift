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
            ChatListView()
                .tabItem {
                    Label("消息", systemImage: "bubble.left.and.bubble.right")
                }
            ProfileView()
                .tabItem {
                    Label("我的", systemImage: "person")
                }
        }
    }
}
