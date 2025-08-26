import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:hive_flutter/hive_flutter.dart';

import 'core/constants/app_constants.dart';
import 'core/theme/app_theme.dart';
import 'core/network/dio_client.dart';
import 'core/storage/storage_service.dart';
import 'shared/providers/app_providers.dart';
import 'features/auth/states/auth_state.dart';
import 'app_router.dart';

void main() async {
  WidgetsFlutterBinding.ensureInitialized();
  
  // 初始化本地存储
  await Hive.initFlutter();
  await StorageService.init();
  
  // 初始化网络客户端
  DioClient.init();
  
  runApp(
    ProviderScope(
      child: YunaiApp(),
    ),
  );
}

class YunaiApp extends ConsumerWidget {
  const YunaiApp({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final router = ref.watch(appRouterProvider);
    final themeMode = ref.watch(themeModeProvider);
    
    return MaterialApp.router(
      title: AppConstants.appName,
      debugShowCheckedModeBanner: false,
      
      // 主题配置
      theme: AppTheme.lightTheme,
      darkTheme: AppTheme.darkTheme,
      themeMode: themeMode,
      
      // 路由配置
      routerConfig: router,
      
      // 国际化配置
      locale: const Locale('zh', 'CN'),
      supportedLocales: const [
        Locale('zh', 'CN'),
        Locale('en', 'US'),
      ],
      
      // 构建器配置
      builder: (context, child) {
        return MediaQuery(
          // 禁用系统字体缩放
          data: MediaQuery.of(context).copyWith(textScaleFactor: 1.0),
          child: child ?? const SizedBox.shrink(),
        );
      },
    );
  }
}

/// 应用启动页
class SplashScreen extends ConsumerStatefulWidget {
  const SplashScreen({super.key});

  @override
  ConsumerState<SplashScreen> createState() => _SplashScreenState();
}

class _SplashScreenState extends ConsumerState<SplashScreen> {
  @override
  void initState() {
    super.initState();
    _initializeApp();
  }

  Future<void> _initializeApp() async {
    try {
      // 检查登录状态
      await ref.read(authStateProvider.notifier).checkAuthStatus();
      
      // 延迟显示启动页
      await Future.delayed(const Duration(seconds: 2));
      
      if (mounted) {
        final authState = ref.read(authStateProvider);
        if (authState.isAuthenticated) {
          // 已登录，跳转到主页
          context.go('/home');
        } else {
          // 未登录，跳转到登录页
          context.go('/auth/login');
        }
      }
    } catch (e) {
      // 初始化失败，跳转到错误页
      if (mounted) {
        context.go('/error');
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Theme.of(context).primaryColor,
      body: Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            // YUNAI Logo
            Container(
              width: 120,
              height: 120,
              decoration: BoxDecoration(
                color: Colors.white,
                borderRadius: BorderRadius.circular(24),
                boxShadow: [
                  BoxShadow(
                    color: Colors.black.withOpacity(0.1),
                    blurRadius: 20,
                    offset: const Offset(0, 8),
                  ),
                ],
              ),
              child: const Icon(
                Icons.smart_toy_outlined,
                size: 60,
                color: Colors.deepPurple,
              ),
            ),
            
            const SizedBox(height: 32),
            
            // 应用名称
            Text(
              AppConstants.appName,
              style: Theme.of(context).textTheme.headlineMedium?.copyWith(
                color: Colors.white,
                fontWeight: FontWeight.bold,
              ),
            ),
            
            const SizedBox(height: 8),
            
            // 应用描述
            Text(
              AppConstants.appDescription,
              style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                color: Colors.white.withOpacity(0.8),
              ),
            ),
            
            const SizedBox(height: 48),
            
            // 加载指示器
            const CircularProgressIndicator(
              valueColor: AlwaysStoppedAnimation<Color>(Colors.white),
            ),
          ],
        ),
      ),
    );
  }
}
