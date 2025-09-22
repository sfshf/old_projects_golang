let user = process.env.NEXT_PUBLIC_USER ?? '';
let password = process.env.NEXT_PUBLIC_PASSWORD ?? '';
const kongAddress = process.env.NEXT_PUBLIC_KONG_ADDRESS ?? '';
let loadedFromLocalStorage = false;

const getPassword = () => {
  // try to load from local storage
  if (
    password === '' &&
    !loadedFromLocalStorage &&
    typeof localStorage !== 'undefined'
  ) {
    loadedFromLocalStorage = true;
    const storedPassword = localStorage.getItem('password');
    if (storedPassword) {
      password = storedPassword;
    }
  }
  return password;
};

const setPassword = (newPassword: string) => {
  // console.log("setPassword", newPassword);
  password = newPassword;
  if (typeof localStorage !== 'undefined' && password !== '') {
    localStorage.setItem('password', password);
  }
};

const shared = {
  getPassword,
  setPassword,
  kongAddress,
};

export default shared;
