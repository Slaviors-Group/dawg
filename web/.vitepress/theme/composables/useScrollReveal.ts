import { ref, onMounted, onUnmounted, type Ref } from 'vue'

export function useScrollReveal(options = { threshold: 0.1, rootMargin: '0px 0px -50px 0px' }) {
  const isRevealed = ref(false)
  const sectionRef = ref<HTMLElement | null>(null)
  let observer: IntersectionObserver | null = null

  onMounted(() => {
    observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          isRevealed.value = true
          observer?.disconnect()
        }
      },
      options
    )
    if (sectionRef.value) observer.observe(sectionRef.value)
  })

  onUnmounted(() => {
    observer?.disconnect()
  })

  return { isRevealed, sectionRef }
}
