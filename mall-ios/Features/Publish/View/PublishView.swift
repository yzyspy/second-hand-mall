import SwiftUI

struct PublishView: View {
    @State private var viewModel = PublishViewModel()

    var body: some View {
        Text(viewModel.title)
    }
}
