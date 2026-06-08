<!--
  TEMPLATE — fill in the {{PLACEHOLDERS}}, have it reviewed by counsel, then host
  it at a public URL and paste that URL into App Store Connect → App Privacy →
  "Privacy Policy URL". This is a starting point, not legal advice.

  Placeholders:
    {{APP_NAME}}        e.g. LIM · Less is More
    {{COMPANY}}         legal entity / developer name
    {{CONTACT_EMAIL}}   privacy contact, e.g. privacy@lim.app
    {{EFFECTIVE_DATE}}  e.g. 2026-06-08
    {{JURISDICTION}}    governing law, e.g. the People's Republic of China / your region
    {{WEBSITE}}         marketing/support site, e.g. https://lim.app
    {{LLM_PROVIDER}}    e.g. Anthropic (Claude) — only if the AI uses a third-party model
-->

# Privacy Policy · 隐私政策

**Effective date / 生效日期: {{EFFECTIVE_DATE}}**

---

## English

{{COMPANY}} (“we”, “us”) operates the {{APP_NAME}} mobile app (the “App”). This
policy explains what we collect, why, and your choices. We follow a “less is
more” principle for data, too: we collect only what we need to give you advice.

### 1. Information we collect

**You provide:**
- **Account:** your email address and a password (stored only as a salted bcrypt
  hash — we never store your plain password).
- **Financial profile (optional but recommended):** monthly income, household
  income, household size, monthly debt/repayments, and a disposable budget. Used
  only to make the AI’s advice fit your situation.
- **Purchase enquiries:** the item, price, category and the reason you enter when
  you ask whether to buy something.

**Created as you use the App:**
- Your decisions (buy / pause / resist), the six-dimension scores and impulse
  index, your cooling-off wishlist, derived statistics (amount saved, streaks,
  growth stage), your chosen app icon, and a per-day count of AI consultations.
- Subscription status (LIM Plus tier and expiry) and a record of transactions.

**Technical:** a sign-in token stored on your device, and a local cache of your
most-recent data so the App works offline. We do **not** use third-party
advertising or analytics SDKs and we do **not** track you across other apps.

### 2. How we use it
To create your account, personalize and generate purchase advice, show your
savings and history, operate the cooling-off reminders, process subscriptions,
secure the service, and comply with law.

### 3. Third parties we share with
We do not sell your data. We share the minimum necessary with:
- **Apple** — to process App Store subscriptions (we receive only a signed,
  verified transaction, not your payment details).
- **{{LLM_PROVIDER}}** *(only if AI analysis is powered by a third-party model)* —
  when you ask for advice, the item, price, reason and the relevant profile
  context may be sent to the model provider to generate the analysis. It is not
  used to train their models where such an option is available to us. See the
  provider’s privacy policy.
- **Service providers** that host our servers, under contract and confidentiality.
- **Legal** — if required by law or to protect rights and safety.

If processing involves transferring data across borders (e.g. to a model provider
in another country), we rely on appropriate safeguards and obtain consent where
required.

### 4. Storage, security & retention
Data is encrypted in transit (HTTPS/TLS). Passwords are bcrypt-hashed; sessions
use signed tokens. We retain your data while your account is active and delete or
anonymize it within a reasonable period after you delete your account, except
where retention is legally required.

### 5. Your rights
Depending on your region (e.g. GDPR, CCPA/CPRA, PIPL) you may access, correct,
export, or delete your data, object to or restrict processing, and withdraw
consent. You can edit your profile in-app, and request deletion or a copy by
emailing {{CONTACT_EMAIL}} — we respond within 30 days.

### 6. Children
The App is rated 4+ but is intended for adults making spending decisions. We do
not knowingly collect data from children under the age required by your region
without verifiable parental consent.

### 7. Changes
We may update this policy; we’ll post the new effective date here and, for
material changes, notify you in-app.

### 8. Contact
{{COMPANY}} · {{CONTACT_EMAIL}} · {{WEBSITE}}. This policy is governed by the laws
of {{JURISDICTION}}.

---

## 简体中文

{{COMPANY}}（以下称“我们”）运营 {{APP_NAME}} 移动应用（以下称“本应用”）。本政策说明
我们收集哪些信息、为何收集，以及你的选择。在数据上，我们同样信奉“少即是多”——只收集
为你提供建议所必需的信息。

### 1. 我们收集的信息

**你主动提供：**
- **账户**：邮箱与密码（仅以加盐 bcrypt 哈希存储，绝不保存明文密码）。
- **财务画像（可选，但建议填写）**：个人月收入、家庭月收入、家庭成员数、每月固定负债/
  月供、可支配预算。仅用于让 AI 的建议更贴合你的实际情况。
- **消费咨询内容**：你在询问“该不该买”时填写的物品、价格、分类与购买理由。

**使用过程中产生：**
- 你的决定（买 / 再想想 / 不买）、六维评分与冲动指数、冷静期心愿单、派生统计（已省金额、
  连续克制、成长阶段）、所选 App 图标，以及每日 AI 咨询次数。
- 订阅状态（LIM Plus 等级与到期时间）与交易记录。

**技术信息**：保存在你设备上的登录令牌，以及用于离线可用的最近数据本地缓存。我们**不**
使用第三方广告或分析 SDK，**不**跨应用追踪你。

### 2. 我们如何使用
用于创建账户、个性化并生成消费建议、展示你的省钱与历史、运行冷静期提醒、处理订阅、保障
服务安全以及遵守法律。

### 3. 第三方共享
我们不出售你的数据，仅在必要范围内共享给：
- **Apple**——处理 App Store 订阅（我们仅收到经签名校验的交易凭证，不接触你的支付信息）。
- **{{LLM_PROVIDER}}**（仅当 AI 分析由第三方大模型提供时）——当你请求建议时，物品、价格、
  理由及相关画像上下文可能发送给模型服务方以生成分析；在我们可选择的范围内，不用于训练
  其模型。详见该服务方隐私政策。
- **服务提供商**——为我们托管服务器，受合同与保密义务约束。
- **法律要求**——在法律要求或为保护权利与安全时。

若处理涉及数据跨境传输（例如发送至位于其他国家/地区的模型服务方），我们将采取适当保护
措施，并在法律要求时取得你的同意。

### 4. 存储、安全与保留
数据传输全程加密（HTTPS/TLS）；密码经 bcrypt 哈希；会话使用签名令牌。账户存续期间我们
保留你的数据，并在你删除账户后的合理期限内删除或匿名化，法律另有要求的除外。

### 5. 你的权利
根据你所在地区（如 GDPR、CCPA/CPRA、个人信息保护法 PIPL），你可能有权访问、更正、导出、
删除你的个人信息，反对或限制处理，并撤回同意。你可在应用内编辑画像；如需删除或获取副本，
请发邮件至 {{CONTACT_EMAIL}}，我们将在 30 天内回复。

### 6. 未成年人
本应用分级为 4+，但面向进行消费决策的成年人。未经可验证的监护人同意，我们不会在明知的
情况下收集你所在地区规定年龄以下儿童的信息。

### 7. 政策变更
我们可能更新本政策；新版生效日期将在此公布，重大变更会在应用内通知你。

### 8. 联系我们
{{COMPANY}} · {{CONTACT_EMAIL}} · {{WEBSITE}}。本政策受 {{JURISDICTION}} 法律管辖。

---

> ⚠️ 本文为模板，非法律意见。上线前请由法律顾问审阅，并据实勾选 App Store Connect 的
> “App 隐私”问卷（收集的数据类型须与本政策一致）。
