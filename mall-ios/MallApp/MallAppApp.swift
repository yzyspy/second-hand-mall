import SwiftUI

@main
struct MallAppApp: App {
    var body: some Scene {
        WindowGroup {
            MainTabView()
        }
    }
}

struct MainTabView: View {
    @State private var selectedTab = 0

    var body: some View {
        TabView(selection: $selectedTab) {
            HomeView()
                .tag(0)
                .tabItem {
                    Image(systemName: "house.fill")
                    Text("首页")
                }

            PublishView()
                .tag(1)
                .tabItem {
                    Image(systemName: "plus.square.fill")
                    Text("发布")
                }

            MessagesView()
                .tag(2)
                .tabItem {
                    Image(systemName: "message.fill")
                    Text("消息")
                }

            ProfileView()
                .tag(3)
                .tabItem {
                    Image(systemName: "person.fill")
                    Text("我的")
                }
        }
        .tint(.mallGreen)
        #if DEBUG
        // 供 UI 调试/截图自动化直达指定 Tab：SIMCTL_CHILD_DEBUG_SELECTED_TAB=<0-3>
        .onAppear {
            if let tabString = ProcessInfo.processInfo.environment["DEBUG_SELECTED_TAB"],
               let tab = Int(tabString) {
                selectedTab = tab
            }
        }
        #endif
    }
}

extension Color {
    /// 品牌绿，对应设计稿中的价格角标 / 选中态颜色
    static let mallGreen = Color(red: 0x07 / 255.0, green: 0xC1 / 255.0, blue: 0x60 / 255.0)
}

#Preview {
    MainTabView()
}
