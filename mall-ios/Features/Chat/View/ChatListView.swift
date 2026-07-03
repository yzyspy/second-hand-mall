import SwiftUI

struct ChatListView: View {
    @State private var viewModel = ChatListViewModel()

    var body: some View {
        Text(viewModel.title)
    }
}
