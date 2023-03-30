import Vue from 'vue'
import VueI18n from 'vue-i18n'
// import messages from '@/locales' // 语言包的地址，随项目本身设置修改

Vue.use(VueI18n)

export const i18n = new VueI18n({
  locale: 'en', 
  fallbackLocale: 'en', // 默认语言设置，当其他语言没有的情况下，使用en作为默认语言
//   messages
})
const loadedLanguages = ['en'] // our default language that is prelaoded 

function setI18nLanguage (lang) {
  i18n.locale = lang
  document.querySelector('html').setAttribute('lang', lang) // 根元素增加lang属性
  return lang
}

export function loadLanguageAsync (lang) {
  if (i18n.locale !== lang) {
    if (!loadedLanguages.includes(lang)) {
      return import(/* webpackChunkName: "lang-[request]" */ `@/locales/${lang}.json`).then(msgs => {
        i18n.setLocaleMessage(lang, msgs)
        loadedLanguages.push(lang)
        return setI18nLanguage(lang)
      })
    } 
    return Promise.resolve(setI18nLanguage(lang))
  }
  return Promise.resolve(lang)
}