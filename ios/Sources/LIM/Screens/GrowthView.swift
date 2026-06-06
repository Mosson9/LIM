import SwiftUI

/// 成长 tab — the savings tree, goal progress, stat tiles, the spent-vs-saved
/// chart, and the most-resisted categories. Backed by GET /stats.
struct GrowthView: View {
    @EnvironmentObject var model: AppModel
    private let goal = 25000

    var body: some View {
        let s = model.stats
        let pct = min(1.0, Double(s.totalSaved) / Double(goal))
        ScrollView {
            VStack(spacing: 14) {
                HStack { Kicker(text: "我的成长", color: Theme.savedDeep); Spacer() }
                    .padding(.top, 58)

                // hero tree
                Card(padding: 24, background: Theme.surface) {
                    VStack(spacing: 4) {
                        GrowthTree(stage: s.treeStage, size: 180)
                        Text("累计省下，种出一棵").font(Theme.sans(12.5)).foregroundColor(Theme.ink3)
                        HStack(alignment: .top, spacing: 2) {
                            Text("¥").font(Theme.display(26)).foregroundColor(Theme.savedDeep)
                            Text(Format.n(s.totalSaved)).font(Theme.display(46)).foregroundColor(Theme.savedDeep)
                        }
                        // goal bar
                        VStack(spacing: 8) {
                            HStack {
                                Text("距离下一阶段「成树」").font(Theme.sans(12.5)).foregroundColor(Theme.ink3)
                                Spacer()
                                Text("\(Int(pct*100))%").font(Theme.sans(12.5, .semibold)).foregroundColor(Theme.savedDeep)
                            }
                            GeometryReader { geo in
                                ZStack(alignment: .leading) {
                                    Capsule().fill(Theme.savedSoft).frame(height: 8)
                                    Capsule().fill(Theme.saved).frame(width: geo.size.width * pct, height: 8)
                                }
                            }.frame(height: 8)
                            HStack {
                                Spacer()
                                Text("还差 \(Format.yuan(max(0, goal - s.totalSaved)))")
                                    .font(Theme.sans(11.5)).foregroundColor(Theme.ink3)
                            }
                        }
                        .padding(.top, 16)
                    }
                    .frame(maxWidth: .infinity)
                }

                // stat tiles
                LazyVGrid(columns: [GridItem(.flexible()), GridItem(.flexible())], spacing: 12) {
                    tile("忍住次数", "\(s.resistCount)", "次", Theme.savedDeep, "leaf")
                    tile("理性购买", "\(s.buyCount)", "次", Theme.ink, "check")
                    tile("连续克制", "\(s.streak)", "天", Theme.indigo, "bolt")
                    tile("克制率", "\(s.restraintRate)", "%", Theme.warn, "target")
                }

                // monthly chart
                Card {
                    VStack(spacing: 18) {
                        HStack {
                            Text("消费 vs 省下").font(Theme.sans(15, .semibold)).foregroundColor(Theme.ink)
                            Spacer()
                            legendDot(Theme.saved, "省下")
                            legendDot(Theme.spent.opacity(0.5), "花掉")
                        }
                        MiniBars(data: s.byMonth, height: 110)
                    }
                }

                // top categories
                if !s.topCats.isEmpty {
                    Card {
                        VStack(alignment: .leading, spacing: 14) {
                            Text("最该克制的类别").font(Theme.sans(15, .semibold)).foregroundColor(Theme.ink)
                            let maxv = s.topCats.first?.saved ?? 1
                            ForEach(s.topCats) { c in
                                HStack(spacing: 12) {
                                    Icon(name: model.category(c.cat).icon, size: 17, color: Color(webHex: c.color))
                                        .frame(width: 34, height: 34)
                                        .background(Color(webHex: model.category(c.cat).soft))
                                        .clipShape(RoundedRectangle(cornerRadius: 10, style: .continuous))
                                    VStack(alignment: .leading, spacing: 6) {
                                        HStack {
                                            Text(c.label).font(Theme.sans(14, .medium)).foregroundColor(Theme.ink)
                                            Spacer()
                                            Text(Format.yuan(c.saved)).font(Theme.sans(13, .semibold)).foregroundColor(Theme.savedDeep)
                                        }
                                        GeometryReader { geo in
                                            ZStack(alignment: .leading) {
                                                Capsule().fill(Theme.paper2).frame(height: 5)
                                                Capsule().fill(Color(webHex: c.color).opacity(0.85))
                                                    .frame(width: geo.size.width * CGFloat(c.saved) / CGFloat(maxv), height: 5)
                                            }
                                        }.frame(height: 5)
                                    }
                                }
                            }
                        }
                    }
                }

                // plus nudge
                Card(background: Theme.ink) {
                    HStack(spacing: 14) {
                        Icon(name: "crown", size: 26, color: Theme.gold)
                        VStack(alignment: .leading, spacing: 2) {
                            Text("解锁专属成长皮肤").font(Theme.sans(14.5, .semibold)).foregroundColor(.white)
                            Text("樱花 · 银杏 · 极光，记录每一次克制")
                                .font(Theme.sans(12.5)).foregroundColor(.white.opacity(0.7))
                        }
                        Spacer()
                        Icon(name: "chevR", size: 18, color: .white.opacity(0.6))
                    }
                }
                .onTapGesture { model.go(.plus) }

                Color.clear.frame(height: 110)
            }
            .padding(.horizontal, 22)
        }
        .background(Theme.paper.ignoresSafeArea())
        .refreshable { await model.refreshAll() }
    }

    private func tile(_ label: String, _ value: String, _ unit: String, _ color: Color, _ icon: String) -> some View {
        Card(padding: 16) {
            VStack(alignment: .leading, spacing: 8) {
                HStack {
                    Text(label).font(Theme.sans(12.5)).foregroundColor(Theme.ink3)
                    Spacer()
                    Icon(name: icon, size: 17, color: color)
                }
                HStack(alignment: .firstTextBaseline, spacing: 2) {
                    Text(value).font(Theme.display(30)).foregroundColor(color)
                    Text(unit).font(Theme.sans(15)).foregroundColor(Theme.ink3)
                }
            }
        }
    }

    private func legendDot(_ color: Color, _ label: String) -> some View {
        HStack(spacing: 5) {
            Circle().fill(color).frame(width: 9, height: 9)
            Text(label).font(Theme.sans(11.5)).foregroundColor(Theme.ink3)
        }
    }
}
