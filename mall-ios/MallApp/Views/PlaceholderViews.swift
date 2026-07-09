import SwiftUI

/// 消息 Tab 的占位页面，后续单独实现

struct MessagesView: View {
    var body: some View {
        NavigationStack {
            placeholder(icon: "message", text: "消息中心（待实现）")
                .navigationTitle("消息")
                .navigationBarTitleDisplayMode(.inline)
        }
    }
}

private func placeholder(icon: String, text: String) -> some View {
    VStack(spacing: 16) {
        Image(systemName: icon)
            .font(.system(size: 48))
            .foregroundStyle(.tertiary)
        Text(text)
            .font(.subheadline)
            .foregroundStyle(.secondary)
    }
    .frame(maxWidth: .infinity, maxHeight: .infinity)
    .background(Color(.systemGray6))
}

#Preview {
    MessagesView()
}
