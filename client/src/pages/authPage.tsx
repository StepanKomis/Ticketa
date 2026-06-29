import './authPage.scss'
import LoginForm from '../components/auth/loginForm'
import RegisterForm from '../components/auth/registerForm'

interface AuthPageProps {
  form: 'login' | 'register'
}

export default function AuthPage({ form }: AuthPageProps) {
  return (
    <div className="authPage">
      {form === 'login' ? <LoginForm /> : <RegisterForm />}
    </div>
  )
}
