import SwiftUI

/// Step ① — the user describes what they want to buy. Feeds into analysis.
struct AskView: View {
    @EnvironmentObject var model: AppModel
    @State private var item = ""
    @State private var price = ""
    @State private var cat = "digital"
    @State private var reason = ""

    private var ready: Bool { !item.trimmingCharacters(in: .whitespaces).isEmpty && Int(price) ?? 0 > 0 }
    private var catList: [Category] { model.categories.filter { $0.id != "other" } }

    var body: some View {
        VStack(spacing: 0) {
            TopBar(title: "问问 LIM", leading: .close)
            ScrollView {
                VStack(alignment: .leading, spacing: 0) {
                    Text("“你心动的是什么？\n一起看看。”")
                        .font(Theme.serif(21)).foregroundColor(Theme.ink)
                        .lineSpacing(6).padding(.top, 8).padding(.bottom, 26)

                    label("想买的东西")
                    textField("例如：索尼降噪耳机", text: $item)

                    label("大概多少钱")
                    HStack(alignment: .firstTextBaseline, spacing: 6) {
                        Text("¥").font(Theme.display(26)).foregroundColor(Theme.ink3)
                        TextField("0", text: $price)
                            .keyboardType(.numberPad)
                            .font(Theme.display(26))
                            .onChange(of: price) { price = $0.filter(\.isNumber) }
                    }
                    .padding(14).background(Theme.surface)
                    .clipShape(RoundedRectangle(cornerRadius: 14, style: .continuous))
                    .overlay(RoundedRectangle(cornerRadius: 14).stroke(Theme.hairline, lineWidth: 1))
                    .padding(.bottom, 18)

                    label("它属于")
                    FlexWrap(spacing: 9, lineSpacing: 9) {
                        ForEach(catList) { c in
                            Button { cat = c.id } label: {
                                Chip(label: c.label, icon: nil, selected: cat == c.id,
                                     tint: Color(webHex: c.color))
                            }
                        }
                    }
                    .padding(.bottom, 20)

                    label("为什么想买它？（可选，越坦白越准）")
                    TextField("直播间看到在打折 / 同事都有了 / 旧的还能用但想换…",
                              text: $reason, axis: .vertical)
                        .lineLimit(3...5)
                        .font(Theme.sans(15))
                        .padding(14).background(Theme.surface)
                        .clipShape(RoundedRectangle(cornerRadius: 14, style: .continuous))
                        .overlay(RoundedRectangle(cornerRadius: 14).stroke(Theme.hairline, lineWidth: 1))
                        .padding(.bottom, 24)
                }
                .padding(.horizontal, 22)
            }

            Button {
                model.go(.analyzing(AnalyzeInput(item: item.trimmingCharacters(in: .whitespaces),
                                                 price: Int(price) ?? 0, cat: cat, reason: reason)))
            } label: {
                HStack(spacing: 8) {
                    Icon(name: "spark", size: 19, color: .white, weight: .semibold)
                    Text("让 LIM 帮我想想")
                }
            }
            .buttonStyle(FilledButtonStyle())
            .disabled(!ready).opacity(ready ? 1 : 0.4)
            .padding(.horizontal, 22).padding(.vertical, 12)
        }
        .background(Theme.paper.ignoresSafeArea())
    }

    private func label(_ t: String) -> some View {
        Text(t).font(Theme.sans(12.5, .semibold)).foregroundColor(Theme.ink3)
            .frame(maxWidth: .infinity, alignment: .leading).padding(.bottom, 8)
    }

    private func textField(_ placeholder: String, text: Binding<String>) -> some View {
        TextField(placeholder, text: text)
            .font(Theme.sans(16))
            .padding(14).background(Theme.surface)
            .clipShape(RoundedRectangle(cornerRadius: 14, style: .continuous))
            .overlay(RoundedRectangle(cornerRadius: 14).stroke(Theme.hairline, lineWidth: 1))
            .padding(.bottom, 18)
    }
}
