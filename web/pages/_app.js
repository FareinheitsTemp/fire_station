import '../styles/main.scss'
import Layout from '../components/Layout'
import ChatWidget from '../components/ChatWidget'

export default function App({ Component, pageProps }) {
  return (
    <Layout>
      <Component {...pageProps} />
      <ChatWidget />
    </Layout>
  )
}
