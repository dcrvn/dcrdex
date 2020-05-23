import Doc from './doc'
import BasePage from './basepage'
import State from './state'
import { postJSON } from './http'
import * as forms from './forms'

const animationLength = 300

var app

export default class SettingsPage extends BasePage {
  constructor (application, body) {
    super()
    app = application
    const page = this.page = Doc.parsePage(body, [
      'darkMode', 'commitHash',
      // Form configure DEX server
      'dexAddrForm', 'dexAddr', 'certFile', 'selectedCert', 'removeCert', 'addCert',
      'submitDEXAddr', 'dexAddrErr',
      // Form confirm DEX registration and pay fee
      'forms', 'confirmRegForm', 'feeDisplay', 'appPass', 'submitConfirm', 'regErr'
    ])
    Doc.bind(page.darkMode, 'click', () => {
      State.dark(page.darkMode.checked)
      if (page.darkMode.checked) {
        document.body.classList.add('dark')
      } else {
        document.body.classList.remove('dark')
      }
    })
    page.commitHash.textContent = app.commitHash.substring(0, 7)
    Doc.bind(page.certFile, 'change', () => this.readCert())
    Doc.bind(page.removeCert, 'click', () => this.resetCert())
    Doc.bind(page.addCert, 'click', () => this.page.certFile.click())

    forms.bind(page.dexAddrForm, page.submitDEXAddr, () => { this.verifyDEX() })
    forms.bind(page.confirmRegForm, page.submitConfirm, () => { this.registerDEX() })
    Doc.bind(page.forms, 'mousedown', e => {
      if (!Doc.mouseInElement(e, this.currentForm)) Doc.hide(page.forms)
    })
  }

  /* showForm shows a modal form with a little animation. */
  async showForm (form) {
    const page = this.page
    this.currentForm = form
    form.style.right = '10000px'
    Doc.show(page.forms, form)
    const shift = (page.forms.offsetWidth + form.offsetWidth) / 2
    await Doc.animate(animationLength, progress => {
      form.style.right = `${(1 - progress) * shift}px`
    }, 'easeOutHard')
    form.style.right = '0px'
  }

  async readCert () {
    const page = this.page
    const files = page.certFile.files
    if (!files.length) return
    page.selectedCert.textContent = files[0].name
    Doc.show(page.removeCert)
    Doc.hide(page.addCert)
  }

  resetCert () {
    const page = this.page
    page.certFile.value = ''
    page.selectedCert.textContent = this.defaultTLSText
    Doc.hide(page.removeCert)
    Doc.show(page.addCert)
  }

  /* Get the reg fees for the DEX. */
  async verifyDEX () {
    const page = this.page
    Doc.hide(page.dexAddrErr)
    const url = page.dexAddr.value
    if (url === '') {
      page.dexAddrErr.textContent = 'URL cannot be empty'
      Doc.show(page.dexAddrErr)
      return
    }

    let cert = ''
    if (page.certFile.value) {
      cert = await page.certFile.files[0].text()
    }

    app.loading(page.dexAddrForm)
    const res = await postJSON('/api/getfee', {
      url: url,
      cert: cert
    })
    app.loaded()
    if (!app.checkResponse(res)) {
      page.dexAddrErr.textContent = res.msg
      Doc.show(page.dexAddrErr)
      return
    }
    this.fee = res.fee

    page.feeDisplay.textContent = Doc.formatCoinValue(res.fee / 1e8)
    await this.showForm(page.confirmRegForm)
  }

  /* Authorize DEX registration. */
  async registerDEX () {
    const page = this.page
    Doc.hide(page.regErr)
    let cert = ''
    if (page.certFile.value) {
      cert = await page.certFile.files[0].text()
    }
    const registration = {
      url: page.dexAddr.value,
      pass: page.appPass.value,
      fee: this.fee,
      cert: cert
    }
    page.appPass.value = ''
    app.loading(page.confirmRegForm)
    const res = await postJSON('/api/register', registration)
    app.loaded()
    if (!app.checkResponse(res)) {
      page.regErr.textContent = res.msg
      Doc.show(page.regErr)
    }
    page.dexAddr.value = ''
    this.resetCert()
    Doc.hide(page.forms)
  }
}
